# dBase Examples

This directory contains comprehensive examples demonstrating various features of the go-dbase library.

## Running Examples

Run all examples:
```bash
make
```

Run specific examples:
```bash
make read write
```

## Available Examples

### Basic Operations
- **[read](./read/)** - Basic file reading and data access patterns
- **[write](./write/)** - Creating and modifying dBase files
- **[create](./create/)** - Building new tables from scratch

### Data Sources
- **[bytes](./bytes/)** - Working with in-memory byte data instead of files
- **[reader](./reader/)** - Using custom io.ReadWriteSeeker implementations
- **[custom](./custom/)** - Advanced custom data source integration

### Advanced Features
- **[search](./search/)** - Finding specific records in tables
- **[database](./database/)** - Working with Visual FoxPro database files (.DBC)
- **[schema](./schema/)** - Analyzing and generating table structure documentation
- **[documentation](./documentation/)** - Generating comprehensive database documentation

## Data Source Comparison

| Example | Data Source | Use Case |
|---------|-------------|----------|
| read, write, create, search | Filesystem (`Filename`) | Standard file operations |
| bytes | Byte arrays (`Data`) | In-memory processing, network data |
| reader | Readers (`Reader`) | Custom I/O control, streams |
| custom | Custom IO (`IO`) | Advanced implementations |

## Key Concepts Demonstrated

### Configuration Validation
All examples show proper configuration with exactly one data source:

```go
// ✅ Valid - filesystem access
config := &dbase.Config{
    Filename: "data.dbf",
}

// ✅ Valid - byte data
config := &dbase.Config{
    Data: dbfBytes,
}

// ❌ Invalid - multiple sources
config := &dbase.Config{
    Filename: "data.dbf",
    Data:     dbfBytes,  // Error!
}
```

### Error Handling
All examples demonstrate proper error handling and resource cleanup:

```go
table, err := dbase.OpenTable(config)
if err != nil {
    log.Fatal(err)
}
defer table.Close() // Always close resources
```

### Data Access Patterns
Examples show different ways to access data:

```go
// Field-by-field access
productID := row.FieldByName("PRODUCTID").GetValue()

// Type-safe access
name, err := row.StringValueByName("NAME")

// Conversion to Go types
var product Product
err := row.ToStruct(&product)

// Map conversion
dataMap, err := row.ToMap()
```

## Test Data

The `test_data/` directory contains sample dBase files used by the examples:

- **table/TEST.DBF** - Sample table with various data types
- **table/TEST.FPT** - Associated memo file
- **database/EXPENSES.DBC** - Visual FoxPro database file
- **database/employees.DBF** - Employee table
- **oms/** - Collection of legacy business system files

## Running Individual Examples

Each example can be run independently:

```bash
cd read && go run read.go
cd bytes && go run bytes.go
cd custom && go run custom.go
```

## Debug Mode

Enable debug logging in any example:

```go
import "os"

// Add at the beginning of main()
dbase.Debug(true, os.Stdout)
```

This will show detailed information about file operations, data parsing, and internal processes.