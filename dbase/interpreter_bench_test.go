package dbase

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// benchRows is the number of rows written into the synthetic table used by the
// read benchmarks. It is large enough for the per-row cost to dominate the
// fixed cost of opening the file, while keeping the setup phase fast.
const benchRows = 10000

// benchColumns builds a 15 column layout that covers every data type reachable
// through Interpret without requiring a memo (FPT) file, so the benchmarks
// measure conversion instead of memo I/O.
func benchColumns(tb testing.TB) []*Column {
	tb.Helper()

	specs := []struct {
		name     string
		dataType DataType
		length   uint8
		decimals uint8
	}{
		{"CODE", Character, 20, 0},
		{"NAME", Character, 50, 0},
		{"DESCR", Character, 100, 0},
		{"UNIT", Character, 10, 0},
		{"TAG", Character, 15, 0},
		{"QTY", Numeric, 12, 0},
		{"PRICE", Numeric, 14, 4},
		{"TOTAL", Numeric, 14, 2},
		{"RATE", Float, 12, 4},
		{"WEIGHT", Double, 0, 0},
		{"AMOUNT", Currency, 0, 0},
		{"ID", Integer, 0, 0},
		{"ACTIVE", Logical, 0, 0},
		{"CREATED", Date, 0, 0},
		{"UPDATED", DateTime, 0, 0},
	}

	columns := make([]*Column, 0, len(specs))
	for _, spec := range specs {
		column, err := NewColumn(spec.name, spec.dataType, spec.length, spec.decimals, false)
		if err != nil {
			tb.Fatalf("creating column %s failed: %v", spec.name, err)
		}
		columns = append(columns, column)
	}
	return columns
}

// benchValues returns a plausible value for every column of the layout above.
func benchValues(i int) map[string]interface{} {
	return map[string]interface{}{
		"CODE":    fmt.Sprintf("ART-%08d", i),
		"NAME":    fmt.Sprintf("Product number %d", i),
		"DESCR":   "A reasonably long description of the product to fill the column",
		"UNIT":    "pcs",
		"TAG":     "tag",
		"QTY":     int64(i % 1000),
		"PRICE":   float64(i%10000) + 0.5,
		"TOTAL":   float64(i%10000) * 1.25,
		"RATE":    float64(i%100) / 3.0,
		"WEIGHT":  float64(i) * 0.125,
		"AMOUNT":  float64(i%10000) + 0.25,
		"ID":      int32(i),
		"ACTIVE":  i%2 == 0,
		"CREATED": time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i%365),
		"UPDATED": time.Date(2024, time.January, 1, 12, 30, 15, 0, time.UTC).Add(time.Duration(i) * time.Minute),
	}
}

// benchTable writes a synthetic table with benchRows rows into a temporary
// directory and returns its path. The repository ships no fixture large enough
// to make per-row costs visible, so the table is generated on the fly using the
// package's own writing support.
//
// The directory is created relative to the working directory and with an
// upper case name on purpose: the Create implementations upper case the whole
// path, so a table cannot be created below a lower case directory such as the
// one returned by testing.TB.TempDir.
func benchTable(tb testing.TB) string {
	tb.Helper()

	dir, err := os.MkdirTemp(".", "BENCHDATA")
	if err != nil {
		tb.Fatalf("creating bench directory failed: %v", err)
	}
	tb.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	path := filepath.Join(dir, "BENCH.DBF")
	file, err := NewTable(
		FoxPro,
		&Config{
			Filename:   path,
			Converter:  NewDefaultConverter(charmap.Windows1250),
			TrimSpaces: true,
		},
		benchColumns(tb),
		0,
		nil,
	)
	if err != nil {
		tb.Fatalf("creating bench table failed: %v", err)
	}

	for i := 0; i < benchRows; i++ {
		row, err := file.RowFromMap(benchValues(i))
		if err != nil {
			tb.Fatalf("building row %d failed: %v", i, err)
		}
		if err := row.Add(); err != nil {
			tb.Fatalf("writing row %d failed: %v", i, err)
		}
	}

	if err := file.Close(); err != nil {
		tb.Fatalf("closing bench table failed: %v", err)
	}
	return path
}

// benchOpen opens the synthetic table for reading.
func benchOpen(tb testing.TB, path string) *File {
	tb.Helper()

	file, err := OpenTable(&Config{Filename: path, TrimSpaces: true})
	if err != nil {
		tb.Fatalf("opening bench table failed: %v", err)
	}
	tb.Cleanup(func() {
		_ = file.Close()
	})
	return file
}

// BenchmarkTableReadAll walks the whole table with Next and Values, which is
// the hot path of the library: Next -> Row -> BytesToRow -> Interpret, once per
// column of every row.
func BenchmarkTableReadAll(b *testing.B) {
	file := benchOpen(b, benchTable(b))

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := file.GoTo(0); err != nil {
			b.Fatalf("seeking to first row failed: %v", err)
		}
		for !file.EOF() {
			row, err := file.Next()
			if err != nil {
				b.Fatalf("reading row failed: %v", err)
			}
			_ = row.Values()
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*benchRows), "ns/row")
}

// BenchmarkBytesToRow isolates the row decoding step from the file I/O by
// converting one pre-read row buffer over and over.
func BenchmarkBytesToRow(b *testing.B) {
	file := benchOpen(b, benchTable(b))

	data, err := file.ReadRow(0)
	if err != nil {
		b.Fatalf("reading row failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := file.BytesToRow(data); err != nil {
			b.Fatalf("converting row failed: %v", err)
		}
	}
}

// BenchmarkInterpret measures Interpret alone, once per column of a single row.
func BenchmarkInterpret(b *testing.B) {
	file := benchOpen(b, benchTable(b))

	data, err := file.ReadRow(0)
	if err != nil {
		b.Fatalf("reading row failed: %v", err)
	}

	columns := file.Columns()
	raws := make([][]byte, len(columns))
	offset := uint16(1)
	for i, column := range columns {
		raws[i] = data[offset : offset+uint16(column.Length)]
		offset += uint16(column.Length)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i, column := range columns {
			if _, err := file.Interpret(raws[i], column); err != nil {
				b.Fatalf("interpreting column %s failed: %v", column.Name(), err)
			}
		}
	}
}

// BenchmarkRepresent measures Represent alone, once per column of a single row.
func BenchmarkRepresent(b *testing.B) {
	file := benchOpen(b, benchTable(b))

	row, err := file.Row()
	if err != nil {
		b.Fatalf("reading row failed: %v", err)
	}
	fields := row.Fields()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, field := range fields {
			if _, err := file.Represent(field, true); err != nil {
				b.Fatalf("representing column %s failed: %v", field.Name(), err)
			}
		}
	}
}
