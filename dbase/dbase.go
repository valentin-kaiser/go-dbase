// Package dbase provides comprehensive tools for reading, writing, and managing dBase-format database files.
//
// This package supports multiple dBase file versions including FoxPro, FoxBase, and dBase III/IV formats.
// It offers cross-platform compatibility with optimized I/O operations for both Unix and Windows systems.
//
// Key Features:
//   - Read and write dBase (.DBF) and memo (.FPT) files
//   - Support for multiple dBase file versions and data types
//   - Flexible data representation (maps, JSON, Go structs)
//   - Character encoding conversion and code page interpretation
//   - Safe concurrent operations with built-in synchronization
//   - Comprehensive error handling with detailed trace information
//   - Navigation and search capabilities within tables
//
// Supported Data Types:
//   - Character (C), Memo (M), Varchar (V) - string data
//   - Numeric (N), Integer (I), Float (F), Currency (Y), Double (B) - numeric data
//   - Date (D), DateTime (T) - temporal data
//   - Logical (L) - boolean data
//   - Blob (W), Varbinary (Q) - binary data
//
// Basic Usage:
//
//	config := &dbase.Config{
//	    Filename: "example.dbf",
//	}
//	table, err := dbase.OpenTable(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer table.Close()
//
//	for !table.EOF() {
//	    row, err := table.Next()
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    // Process row data
//	}
//
// Common use cases include migrating legacy dBase systems, data conversion and integration,
// interfacing with legacy applications, and building tools for dBase file manipulation.
package dbase

// Config is a struct containing the configuration for opening a Foxpro/dbase databse or table.
// The filename is mandatory.
//
// The other fields are optional and are false by default.
// If Converter and InterpretCodePage are both not set the package will try to interpret the code page mark.
// To open untested files set Untested to true. Tested files are defined in the constants.go file.
type Config struct {
	Filename                          string            // The filename of the DBF file.
	Converter                         EncodingConverter // The encoding converter to use.
	Exclusive                         bool              // If true the file is opened in exclusive mode.
	Untested                          bool              // If true the file version is not checked.
	TrimSpaces                        bool              // If true, spaces are trimmed from the start and end of string values.
	CollapseSpaces                    bool              // If true, any length of spaces is replaced by a single space.
	DisableConvertFilenameUnderscores bool              // If false underscores in the table filename are converted to spaces.
	ReadOnly                          bool              // If true the file is opened in read-only mode.
	WriteLock                         bool              // Whether or not the write operations should lock the record
	ValidateCodePage                  bool              // Whether or not the code page mark should be validated.
	InterpretCodePage                 bool              // Whether or not the code page mark should be interpreted. Ignores the defined converter.
	IO                                IO                // The IO interface to use.
}

// Modification allows to change the column name or value type of a column when reading the table
// The TrimSpaces option is only used for a specific column, if the general TrimSpaces option in the config is false.
type Modification struct {
	TrimSpaces  bool                                   // Trim spaces from string values
	Convert     func(interface{}) (interface{}, error) // Conversion function to convert the value
	ExternalKey string                                 // External key to use for the column
}
