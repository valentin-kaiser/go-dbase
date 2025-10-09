package dbase

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestBytesReadWriteSeeker(t *testing.T) {
	// Test creating from nil data
	rwSeeker := NewBytesReadWriteSeeker(nil)
	if rwSeeker != nil {
		t.Error("Expected nil when creating from nil data")
	}

	// Test creating from empty data
	rwSeeker = NewBytesReadWriteSeeker([]byte{})
	if rwSeeker == nil {
		t.Error("Expected non-nil when creating from empty data")
	}

	// Test creating from actual data
	testData := []byte("Hello, World!")
	rwSeeker = NewBytesReadWriteSeeker(testData)
	if rwSeeker == nil {
		t.Error("Expected non-nil when creating from test data")
	}

	// Test reading
	buffer := make([]byte, 5)
	n, err := rwSeeker.Read(buffer)
	if err != nil {
		t.Errorf("Unexpected error reading: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected to read 5 bytes, got %d", n)
	}
	if string(buffer) != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", string(buffer))
	}

	// Test seeking
	pos, err := rwSeeker.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf("Unexpected error seeking: %v", err)
	}
	if pos != 0 {
		t.Errorf("Expected position 0, got %d", pos)
	}

	// Test writing
	n, err = rwSeeker.Write([]byte("Hi"))
	if err != nil {
		t.Errorf("Unexpected error writing: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected to write 2 bytes, got %d", n)
	}

	// Test size and data
	if rwSeeker.Size() != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), rwSeeker.Size())
	}
}

func TestOpenTableFromBytes(t *testing.T) {
	// Check if test file exists
	testFile := "../examples/test_data/table/TEST.DBF"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping test")
	}

	// Read test data from filesystem
	dbfData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test DBF file: %v", err)
	}

	memoFile := "../examples/test_data/table/TEST.FPT"
	var memoData []byte
	if _, err := os.Stat(memoFile); err == nil {
		memoData, _ = os.ReadFile(memoFile)
	}

	// Test opening table from bytes
	table, err := OpenTable(&Config{
		Data:       dbfData,
		MemoData:   memoData,
		TrimSpaces: true,
	})
	if err != nil {
		t.Fatalf("Failed to open table from bytes: %v", err)
	}
	defer table.Close()

	// Verify table properties
	if table.TableName() == "" {
		t.Error("Table name should not be empty")
	}

	if table.RowsCount() == 0 {
		t.Error("Table should have rows")
	}

	if table.ColumnsCount() == 0 {
		t.Error("Table should have columns")
	}

	// Test reading rows
	if !table.BOF() {
		t.Error("Table should be at beginning of file initially")
	}

	if table.EOF() {
		t.Error("Table should not be at end of file initially")
	}

	// Read first row
	row, err := table.Next()
	if err != nil {
		t.Fatalf("Failed to read first row: %v", err)
	}

	if len(row.Values()) == 0 {
		t.Error("Row should have values")
	}
}

func TestOpenTableFromReader(t *testing.T) {
	// Check if test file exists
	testFile := "../examples/test_data/table/TEST.DBF"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping test")
	}

	// Open test files
	dbfFile, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test DBF file: %v", err)
	}
	defer dbfFile.Close()

	memoFile := "../examples/test_data/table/TEST.FPT"
	var memoReader *os.File
	if _, err := os.Stat(memoFile); err == nil {
		memoReader, _ = os.Open(memoFile)
		if memoReader != nil {
			defer memoReader.Close()
		}
	}

	// Test opening table from readers
	table, err := OpenTable(&Config{
		Reader:     dbfFile,
		MemoReader: memoReader,
		TrimSpaces: true,
	})
	if err != nil {
		t.Fatalf("Failed to open table from readers: %v", err)
	}
	defer table.Close()

	// Verify table properties
	if table.TableName() == "" {
		t.Error("Table name should not be empty")
	}

	if table.RowsCount() == 0 {
		t.Error("Table should have rows")
	}

	// Read first row
	row, err := table.Next()
	if err != nil {
		t.Fatalf("Failed to read first row: %v", err)
	}

	if len(row.Values()) == 0 {
		t.Error("Row should have values")
	}
}

func TestOpenTableFromBytesReader(t *testing.T) {
	// Check if test file exists
	testFile := "../examples/test_data/table/TEST.DBF"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping test")
	}

	// Read test data
	dbfData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test DBF file: %v", err)
	}

	// Read memo data if it exists
	memoFile := "../examples/test_data/table/TEST.FPT"
	var memoReader *BytesReadWriteSeeker
	if _, err := os.Stat(memoFile); err == nil {
		memoData, err := os.ReadFile(memoFile)
		if err == nil {
			memoReader = NewBytesReadWriteSeeker(memoData)
		}
	}

	// Use our BytesReadWriteSeeker wrapper
	dbfReader := NewBytesReadWriteSeeker(dbfData)
	if dbfReader == nil {
		t.Fatal("Failed to create BytesReadWriteSeeker")
	}

	// Test opening table from our custom reader
	table, err := OpenTable(&Config{
		Reader:     dbfReader,
		MemoReader: memoReader,
		TrimSpaces: true,
	})
	if err != nil {
		t.Fatalf("Failed to open table from BytesReadWriteSeeker: %v", err)
	}
	defer table.Close()

	// Verify table properties
	if table.TableName() == "" {
		t.Error("Table name should not be empty")
	}

	if table.RowsCount() == 0 {
		t.Error("Table should have rows")
	}
}

func TestOpenDatabaseFromBytes(t *testing.T) {
	// Check if database file exists
	dbcFile := "../examples/test_data/database/EXPENSES.DBC"
	if _, err := os.Stat(dbcFile); os.IsNotExist(err) {
		t.Skip("Test database file not found, skipping test")
	}

	// Read database file
	dbcData, err := os.ReadFile(dbcFile)
	if err != nil {
		t.Fatalf("Failed to read DBC file: %v", err)
	}

	// Read database memo file (DCT)
	dctFile := "../examples/test_data/database/EXPENSES.DCT"
	var dctData []byte
	if _, err := os.Stat(dctFile); err == nil {
		dctData, err = os.ReadFile(dctFile)
		if err != nil {
			t.Fatalf("Failed to read DCT file: %v", err)
		}
	}

	// Create table provider
	tableProvider := func(tableName string) ([]byte, []byte, error) {
		dbfPath := "../examples/test_data/database/" + tableName + ".dbf"
		memoPath := "../examples/test_data/database/" + tableName + ".fpt"

		dbfData, err := os.ReadFile(dbfPath)
		if err != nil {
			// Return nil for missing tables (they might not all exist)
			return nil, nil, nil
		}

		var memoData []byte
		if _, err := os.Stat(memoPath); err == nil {
			memoData, _ = os.ReadFile(memoPath)
		}

		return dbfData, memoData, nil
	}

	// Test opening database from bytes
	db, err := OpenDatabase(&Config{
		Data:          dbcData,
		MemoData:      dctData,
		TableProvider: tableProvider,
		TrimSpaces:    true,
	})
	if err != nil {
		t.Fatalf("Failed to open database from bytes: %v", err)
	}
	defer db.Close()

	// Verify database properties
	tables := db.Tables()
	if len(tables) == 0 {
		t.Error("Database should have tables")
	}

	tableNames := db.Names()
	if len(tableNames) == 0 {
		t.Error("Database should have table names")
	}

	schema := db.Schema()
	if len(schema) == 0 {
		t.Error("Database should have schema")
	}
}

func TestConfigValidation(t *testing.T) {
	// Test with nil config
	_, err := OpenTable(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test with missing data source
	_, err = OpenTable(&Config{
		TrimSpaces: true,
	})
	if err == nil {
		t.Error("Expected error for missing data source")
	}
	if !strings.Contains(err.Error(), "missing filename, data, or reader") {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Test with empty byte data
	_, err = OpenTable(&Config{
		Data: []byte{},
	})
	if err == nil {
		t.Error("Expected error for empty data")
	}

	// Test with nil reader
	_, err = OpenTable(&Config{
		Reader: nil,
	})
	if err == nil {
		t.Error("Expected error for nil reader")
	}
}

func TestOpenDatabaseConfigValidation(t *testing.T) {
	// Test database without table provider when using bytes
	_, err := OpenDatabase(&Config{
		Data: []byte("dummy"),
	})
	if err == nil {
		t.Error("Expected error when opening database with bytes but no table provider")
	}

	// Test database without table reader provider when using readers
	mockReader := NewBytesReadWriteSeeker([]byte("dummy"))
	_, err = OpenDatabase(&Config{
		Reader: mockReader,
	})
	if err == nil {
		t.Error("Expected error when opening database with reader but no table reader provider")
	}
}
