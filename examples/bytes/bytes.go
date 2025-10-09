package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/valentin-kaiser/go-dbase/dbase"
)

func main() {
	// Open debug log file so we see what's going on
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	dbase.Debug(true, io.MultiWriter(os.Stdout, f))

	// Example 1: Opening a table from byte data
	fmt.Println("=== Example 1: Opening table from bytes ===")
	dbfData, err := os.ReadFile("../test_data/table/TEST.DBF")
	if err != nil {
		log.Printf("Failed to read DBF file: %v", err)
		return
	}

	memoData, err := os.ReadFile("../test_data/table/TEST.FPT")
	if err != nil {
		log.Printf("Warning: Failed to read memo file: %v", err)
		memoData = nil // Continue without memo data
	}

	table, err := dbase.OpenTable(&dbase.Config{
		Data:       dbfData,
		MemoData:   memoData,
		TrimSpaces: true,
	})
	if err != nil {
		log.Fatalf("Failed to open table from bytes: %v", err)
	}
	defer table.Close()

	fmt.Printf("Table name: %s\n", table.TableName())
	fmt.Printf("Records: %d\n", table.RowsCount())
	fmt.Printf("Columns: %d\n", table.ColumnsCount())

	// Read first few rows
	for i := 0; i < 3 && !table.EOF(); i++ {
		row, err := table.Next()
		if err != nil {
			log.Printf("Error reading row: %v", err)
			break
		}
		fmt.Printf("Row %d: %v\n", i+1, row.Values())
	}

	// Example 2: Opening a table from readers
	fmt.Println("\n=== Example 2: Opening table from readers ===")
	dbfFile, err := os.Open("../test_data/table/TEST.DBF")
	if err != nil {
		log.Printf("Failed to open DBF file: %v", err)
		return
	}
	defer dbfFile.Close()

	memoFile, err := os.Open("../test_data/table/TEST.FPT")
	if err != nil {
		log.Printf("Warning: Failed to open memo file: %v", err)
		memoFile = nil // Continue without memo file
	}
	if memoFile != nil {
		defer memoFile.Close()
	}

	table2, err := dbase.OpenTable(&dbase.Config{
		Reader:     dbfFile,
		MemoReader: memoFile,
		TrimSpaces: true,
	})
	if err != nil {
		log.Fatalf("Failed to open table from readers: %v", err)
	}
	defer table2.Close()

	fmt.Printf("Table name: %s\n", table2.TableName())
	fmt.Printf("Records: %d\n", table2.RowsCount())

	// Example 3: Opening a database from byte data with table provider
	fmt.Println("\n=== Example 3: Opening database from bytes ===")

	// Check if database file exists
	if _, err := os.Stat("../test_data/database/EXPENSES.DBC"); os.IsNotExist(err) {
		fmt.Println("Database test file not found, skipping database example")
		return
	}

	dbcData, err := os.ReadFile("../test_data/database/EXPENSES.DBC")
	if err != nil {
		log.Printf("Failed to read DBC file: %v", err)
		return
	}

	// Read the DCT memo file if it exists
	var dctData []byte
	if _, err := os.Stat("../test_data/database/EXPENSES.DCT"); err == nil {
		dctData, _ = os.ReadFile("../test_data/database/EXPENSES.DCT")
	}

	// Create a table provider function that reads table files from disk
	tableProvider := func(tableName string) ([]byte, []byte, error) {
		// Try both the name as-is and with underscores replaced by spaces
		possibleNames := []string{
			tableName,
			strings.ReplaceAll(tableName, "_", " "),
		}

		for _, name := range possibleNames {
			dbfPath := fmt.Sprintf("../test_data/database/%s.dbf", name)
			memoPath := fmt.Sprintf("../test_data/database/%s.fpt", name)

			dbfData, err := os.ReadFile(dbfPath)
			if err != nil {
				continue // Try next possible name
			}

			var memoData []byte
			if _, err := os.Stat(memoPath); err == nil {
				memoData, _ = os.ReadFile(memoPath)
			}

			return dbfData, memoData, nil
		}

		return nil, nil, fmt.Errorf("table file not found for %s", tableName)
	}

	db, err := dbase.OpenDatabase(&dbase.Config{
		Data:          dbcData,
		MemoData:      dctData,
		TableProvider: tableProvider,
		TrimSpaces:    true,
	})
	if err != nil {
		log.Fatalf("Failed to open database from bytes: %v", err)
	}
	defer db.Close()

	fmt.Printf("Database tables: %v\n", db.Names())

	// Access a table from the database
	tables := db.Tables()
	for name, table := range tables {
		fmt.Printf("Table %s has %d records\n", name, table.RowsCount())
		break // Just show the first table
	}

	fmt.Println("\n=== All examples completed successfully! ===")
}
