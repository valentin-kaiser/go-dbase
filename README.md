<p align="center">
  <img src="go-dbase.png" width="365">
</p>

# Microsoft Visual FoxPro / dBase Library for Go

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/Valentin-Kaiser/go-dbase)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://github.com/Valentin-Kaiser/go-dbase/blob/main/LICENSE)
[![Tests](https://github.com/Valentin-Kaiser/go-dbase/workflows/Tests/badge.svg)](https://github.com/Valentin-Kaiser/go-dbase)
[![Linters](https://github.com/Valentin-Kaiser/go-dbase/workflows/Linters/badge.svg)](https://github.com/Valentin-Kaiser/go-dbase)
[![CodeQL](https://github.com/Valentin-Kaiser/go-dbase/workflows/CodeQL/badge.svg)](https://github.com/Valentin-Kaiser/go-dbase)
[![Examples](https://github.com/Valentin-Kaiser/go-dbase/workflows/Examples/badge.svg)](https://github.com/Valentin-Kaiser/go-dbase)
[![Go Report](https://goreportcard.com/badge/github.com/Valentin-Kaiser/go-dbase)](https://goreportcard.com/report/github.com/Valentin-Kaiser/go-dbase)

**A comprehensive Golang package for reading, writing, and managing FoxPro dBase table and memo files.**

## Overview

This package provides comprehensive tools for working with dBase-format database files (.DBF) and their associated memo files (.FPT). It offers cross-platform compatibility with optimized I/O operations for both Unix and Windows systems, flexible data representation, and safe concurrent operations.

### Key Features

- 📁 **Full dBase Support**: Read and write .DBF tables and .FPT memo files
- 🔄 **Multiple File Versions**: Support for FoxPro, FoxBase, and dBase III/IV formats  
- 🌐 **Encoding Support**: 13+ character encodings with automatic code page detection
- 🔒 **Concurrent Safe**: Built-in synchronization for multi-threaded applications
- 📊 **Flexible Output**: Convert to Go structs, JSON, maps, or native types
- 🔍 **Advanced Features**: Search, navigation, exclusive file access, and table creation
- ⚡ **Memory Efficient**: Streaming reads without loading entire files into memory
- 🛠️ **Developer Friendly**: Comprehensive error handling and debugging support

### Use Cases

- **Legacy System Migration**: Modernize applications that rely on dBase files
- **Data Integration**: Import/export data from legacy business systems
- **File Conversion**: Convert dBase files to modern database formats
- **Backup & Recovery**: Create tools for dBase file manipulation and repair
- **Business Intelligence**: Extract data from legacy ERP/CRM systems

## Quick Start

### Installation

```bash
go get github.com/Valentin-Kaiser/go-dbase@latest
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/Valentin-Kaiser/go-dbase/dbase"
)

func main() {
    // Open a dBase file
    config := &dbase.Config{
        Filename: "example.dbf",
    }
    
    table, err := dbase.OpenTable(config)
    if err != nil {
        log.Fatal(err)
    }
    defer table.Close()
    
    // Read all rows
    for !table.EOF() {
        row, err := table.Next()
        if err != nil {
            log.Fatal(err)
        }
        
        // Access data by column name
        name, _ := row.StringValueByName("NAME")
        age, _ := row.IntValueByName("AGE")
        
        fmt.Printf("Name: %s, Age: %d\n", name, age)
    }
    
    // Convert row to map
    row, _ := table.Row()
    dataMap := row.ToMap()
    fmt.Printf("Row as map: %+v\n", dataMap)
    
    // Convert row to JSON
    jsonData, _ := row.ToJSON()
    fmt.Printf("Row as JSON: %s\n", jsonData)
}
```

## Features Comparison

Comparison with other popular Go dBase libraries:

| Feature | [go-dbase](https://github.com/Valentin-Kaiser/go-dbase) | [go-dbf](https://github.com/LindsayBradford/go-dbf) | [go-foxpro-dbf](https://github.com/SebastiaanKlippert/go-foxpro-dbf) | 
| --- | :---: | :---: | :---: |
| **File Operations** |
| Read .DBF files | ✅ | ✅ | ✅ |
| Write .DBF files | ✅ | ✅ | ❌ |
| Read .FPT memo files | ✅ | ❌ | ✅ |
| Write .FPT memo files | ✅ | ❌ | ❌ |
| **Data Features** |
| Full data type support | ✅ | ❌ | ❌ |
| Character encoding support | ✅ (13+ encodings) | ✅[*](https://github.com/LindsayBradford/go-dbf/issues/3) | ✅ (extensible) |
| Automatic code page detection | ✅ | ❌ | ❌ |
| **Data Conversion** |
| Go struct mapping | ✅ | ❌ | ✅ |
| JSON conversion | ✅ | ❌ | ✅ |
| Map conversion | ✅ | ❌ | ✅ |
| **Advanced Features** |
| Concurrent access safety | ✅ | ❌ | ❌ |
| Exclusive file locking | ✅ | ❌ | ❌ |
| Search functionality | ✅ | ❌ | ❌ |
| Table creation | ✅ | ❌ | ❌ |
| Database operations | ✅ | ❌ | ❌ |
| Memory efficiency | ✅ | ❌ | ✅ |

### Technical Advantages

**🔒 Concurrent Safety**: Built-in mutex locks ensure safe operations in multi-threaded environments.

**⚡ Memory Efficiency**: Streaming approach reads only required file positions instead of loading entire files into memory, enabling processing of large files with minimal RAM usage.

**🔐 Exclusive Access**: Support for exclusive file locking during write operations prevents data corruption from concurrent access.

**🌐 Encoding Intelligence**: Automatic code page detection and conversion with support for 13+ character encodings.

**🛠️ Developer Experience**: Comprehensive error handling with detailed trace information and extensive helper methods.

## Supported Data Types

All column values are returned as `interface{}` with helper methods for type-safe conversion.

| dBase Type | Type Name | Go Type | Description |
|:----------:|-----------|---------|-------------|
| **Text Data** |
| C | Character | `string` | Fixed-length text fields |
| M | Memo | `string` | Variable-length text in .FPT file |
| V | Varchar | `string` | Variable-length text |
| **Numeric Data** |
| N | Numeric (no decimals) | `int64` | Integer numbers |
| N | Numeric (with decimals) | `float64` | Decimal numbers |
| I | Integer | `int32` | 32-bit signed integers |
| F | Float | `float64` | Floating-point numbers |
| Y | Currency | `float64` | Currency/money values |
| B | Double | `float64` | Double-precision floats |
| **Date/Time Data** |
| D | Date | `time.Time` | Date values (YYYYMMDD) |
| T | DateTime | `time.Time` | Date and time values |
| **Other Data** |
| L | Logical | `bool` | Boolean values (T/F) |
| W | Blob | `[]byte` | Binary large objects |
| Q | Varbinary | `[]byte` | Variable-length binary |
| G | General | `[]byte` | General/OLE objects |
| P | Picture | `[]byte` | Picture/image data |

### Type Conversion Examples

```go
// Type-safe value retrieval
name, err := row.StringValueByName("NAME")        // string
age, err := row.IntValueByName("AGE")             // int64  
salary, err := row.FloatValueByName("SALARY")     // float64
active, err := row.BoolValueByName("ACTIVE")      // bool
hired, err := row.TimeValueByName("HIRE_DATE")    // time.Time
photo, err := row.BytesValueByName("PHOTO")       // []byte

// Panic versions (for when you're sure the field exists)
name := row.MustStringValueByName("NAME")
age := row.MustIntValueByName("AGE")
```

> 📖 **Reference**: [Microsoft Visual Studio FoxPro Data Types](https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/74zkxe2k(v=vs.80))

> **Note**: Need additional column types? Please [open an issue](https://github.com/Valentin-Kaiser/go-dbase/issues) or submit a pull request.

## Character Encoding Support

Automatic detection and conversion of 13+ character encodings with UTF-8 as the standard:

| Code Page | Platform | Identifier | Description |
|-----------|----------|:----------:|-------------|
| 437 | U.S. MS-DOS | `0x01` | Original IBM PC character set |
| 850 | International MS-DOS | `0x02` | Western European |
| 852 | Eastern European MS-DOS | `0x64` | Central/Eastern European |
| 865 | Nordic MS-DOS | `0x66` | Nordic countries |
| 866 | Russian MS-DOS | `0x65` | Cyrillic script |
| 874 | Thai Windows | `0x7C` | Thai script |
| 1250 | Central European Windows | `0xC8` | Central European |
| 1251 | Russian Windows | `0xC9` | Cyrillic Windows |
| 1252 | Windows ANSI | `0x03` | Western European Windows |
| 1253 | Greek Windows | `0xCB` | Greek script |
| 1254 | Turkish Windows | `0xCA` | Turkish script |
| 1255 | Hebrew Windows | `0x7D` | Hebrew script |
| 1256 | Arabic Windows | `0x7E` | Arabic script |

### Encoding Examples

```go
// Automatic encoding detection (recommended)
config := &dbase.Config{
    Filename:          "data.dbf",
    InterpretCodePage: true,  // Auto-detect from file
}

// Manual encoding specification
config := &dbase.Config{
    Filename:  "data.dbf", 
    Converter: dbase.ConverterFromCodePage(0x03), // Windows-1252
}

// Custom encoding registration
import "golang.org/x/text/encoding/charmap"
dbase.RegisterCustomEncoding(0x99, charmap.ISO8859_15)
```

> All encodings are automatically converted to/from UTF-8 for seamless Go integration.

## Advanced Examples

### Creating a New Table

```go
// Define columns
columns := []*dbase.Column{
    dbase.NewColumn("ID", dbase.Integer, 0, 0, false),
    dbase.NewColumn("NAME", dbase.Character, 50, 0, false),
    dbase.NewColumn("SALARY", dbase.Numeric, 10, 2, false),
    dbase.NewColumn("ACTIVE", dbase.Logical, 0, 0, false),
}

config := &dbase.Config{Filename: "employees.dbf"}
table, err := dbase.NewTable(dbase.FoxPro, config, columns, 0, nil)
if err != nil {
    log.Fatal(err)
}
defer table.Close()

// Create and add a new row
row := table.NewRow()
row.SetValueByName("ID", 1)
row.SetValueByName("NAME", "John Doe")
row.SetValueByName("SALARY", 75000.50)
row.SetValueByName("ACTIVE", true)

err = table.AppendRow(row)
if err != nil {
    log.Fatal(err)
}
```

### Working with Memo Fields

```go
// Reading memo data
memoText, err := row.StringValueByName("DESCRIPTION")  // Text memo
memoBytes, err := row.BytesValueByName("DOCUMENT")     // Binary memo

// Writing memo data
row := table.NewRow()
row.SetValueByName("DESCRIPTION", "Long text content...")
row.SetValueByName("DOCUMENT", []byte{0x89, 0x50, 0x4E, 0x47}) // Binary data
```

### Search and Navigation

```go
// Search for specific values
field := &dbase.Field{Name: "STATUS", Value: "ACTIVE"}
results, err := table.Search(field, true) // exactMatch = true

// Navigation
table.GoTo(10)           // Go to specific record
table.Skip(5)            // Skip 5 records forward
table.Skip(-3)           // Skip 3 records backward

// Check position
if table.EOF() {
    fmt.Println("At end of file")
}
if table.BOF() {
    fmt.Println("Before first record")
}
```

### Convert to Different Formats

```go
// Convert row to Go struct
type Employee struct {
    ID     int32   `dbase:"ID"`
    Name   string  `dbase:"NAME"`
    Salary float64 `dbase:"SALARY"`
    Active bool    `dbase:"ACTIVE"`
}

var emp Employee
err := row.ToStruct(&emp)

// Convert to JSON
jsonBytes, err := row.ToJSON()

// Convert to map
dataMap := row.ToMap()
```

### Error Handling and Debugging

```go
// Enable debug mode
dbase.Debug(true, os.Stdout)

// Comprehensive error handling
if err != nil {
    if dbaseErr, ok := err.(dbase.Error); ok {
        fmt.Printf("dBase error: %s\n", dbaseErr.Error())
        // Access error details and trace
    }
}
```

### Working with Large Files

```go
config := &dbase.Config{
    Filename:    "large_file.dbf",
    ReadOnly:    true,        // Read-only for better performance
    TrimSpaces:  true,        // Automatically trim string values
}

table, err := dbase.OpenTable(config)
if err != nil {
    log.Fatal(err)
}
defer table.Close()

// Process in batches
batchSize := 1000
processed := 0

for !table.EOF() {
    batch := make([]*dbase.Row, 0, batchSize)
    
    // Collect batch
    for i := 0; i < batchSize && !table.EOF(); i++ {
        row, err := table.Next()
        if err != nil {
            log.Printf("Error reading row %d: %v", processed+i, err)
            continue
        }
        batch = append(batch, row)
    }
    
    // Process batch
    processBatch(batch)
    processed += len(batch)
    fmt.Printf("Processed %d records\n", processed)
}
```

## Complete Examples

Explore comprehensive examples in the [examples](./examples/) directory:

- 📖 **[Reading Files](./examples/read/read.go)** - Basic file reading and data access
- ✏️ **[Writing Files](./examples/write/write.go)** - Creating and modifying dBase files  
- 🔍 **[Search Operations](./examples/search/search.go)** - Finding specific records
- 🆕 **[Table Creation](./examples/create/create.go)** - Building new tables from scratch
- 📊 **[Database Export](./examples/database/export.go)** - Converting to modern formats
- 📋 **[Documentation](./examples/documentation/documentation.go)** - Generating table documentation
- 🗂️ **[Schema Analysis](./examples/schema/schema.go)** - Analyzing table structures

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

> ⚠️ **Important**: This library is designed for working with **existing** dBase files and legacy system integration. While it supports creating new tables, it should not be used to develop new applications that rely on dBase as the primary database format. Consider modern database solutions for new projects.
