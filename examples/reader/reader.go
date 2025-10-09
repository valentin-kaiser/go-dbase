package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/valentin-kaiser/go-dbase/dbase"
)

func main() {
	// Open debug log file so we see what's going on
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	dbase.Debug(true, io.MultiWriter(os.Stdout, f))

	dbfFile, err := os.OpenFile("../test_data/table/TEST.DBF", os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("Error opening DBF file:", err)
	}

	memoFile, err := os.OpenFile("../test_data/table/TEST.FPT", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Note: No memo file found, continuing without memo data")
	}

	// Configure dBase to use readers instead of filesystem
	config := &dbase.Config{
		Reader:            dbfFile,  // io.ReadWriteSeeker for DBF file
		MemoReader:        memoFile, // io.ReadWriteSeeker for memo file (optional)
		TrimSpaces:        true,
		InterpretCodePage: false,
	}

	table, err := dbase.OpenTable(config)
	if err != nil {
		log.Fatal("Error opening table from readers:", err)
	}
	defer func() {
		table.Close()
		dbfFile.Close()
		if memoFile != nil {
			memoFile.Close()
		}
	}()

	fmt.Printf(
		"Last modified: %v Columns count: %v Record count: %v File size: %v \n",
		table.Header().Modified(0),
		table.Header().ColumnsCount(),
		table.Header().RecordsCount(),
		table.Header().FileSize(),
	)

	// Display column information
	fmt.Println("Columns:")
	for i, column := range table.Columns() {
		fmt.Printf("  %d. %s [%s] - Length: %d\n",
			i+1, column.Name(), column.Type(), column.Length)
	}
}
