package dbase

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type Database struct {
	file   *File
	tables map[string]*File
}

// OpenDatabase opens a dbase/foxpro database file and all related tables.
// You can provide either:
//   - Filename: path to DBC file on filesystem
//   - Data: DBC file content as bytes (with TableProvider for related tables)
//   - Reader: DBC file content as io.ReadWriteSeeker (with TableReaderProvider)
func OpenDatabase(config *Config) (*Database, error) {
	if config == nil {
		return nil, NewError("missing dbase configuration")
	}

	if !(config.Data != nil || config.Reader != nil) {
		if len(strings.TrimSpace(config.Filename)) == 0 {
			return nil, NewError("missing dbase filename")
		}
		if strings.ToUpper(filepath.Ext(config.Filename)) != string(DBC) {
			return nil, NewError("invalid dbase filename").Details(fmt.Errorf("file extension must be %v", DBC))
		}
		debugf("Opening database: %v", config.Filename)
	}

	if config.Data != nil || config.Reader != nil {
		if config.TableProvider == nil && config.TableReaderProvider == nil {
			return nil, NewError("when using Data or Reader for database, you must provide TableProvider or TableReaderProvider")
		}
		debugf("Opening database from byte/reader data")
	}

	databaseTable, err := OpenTable(config)
	if err != nil {
		return nil, WrapError(err)
	}
	// Search by all records where object type is table
	typeField, err := databaseTable.NewFieldByName("OBJECTTYPE", "Table")
	if err != nil {
		return nil, WrapError(err)
	}
	rows, err := databaseTable.Search(typeField, true)
	if err != nil {
		return nil, WrapError(err)
	}
	// Try to load the table files
	tables := make(map[string]*File, 0)
	for _, row := range rows {
		objectName, err := row.ValueByName("OBJECTNAME")
		if err != nil {
			return nil, WrapError(err)
		}
		tableName, ok := objectName.(string)
		if !ok {
			return nil, NewError("table name is not a string")
		}
		tableName = strings.Trim(tableName, " ")
		if tableName == "" {
			continue
		}
		debugf("Found table: %v in database", tableName)

		var tableConfig *Config

		// Check if we're using byte/reader data sources
		if config.Data != nil || config.Reader != nil {
			var err error
			tableConfig, err = buildTableConfig(config, tableName)
			if err != nil {
				return nil, err
			}
			if tableConfig == nil {
				continue // Skip if no data/reader provided
			}
		}

		if config.Data == nil && config.Reader == nil {
			// Use filesystem access
			tablePath := path.Join(filepath.Dir(config.Filename), tableName+string(DBF))
			// Replace underscores with spaces
			if !config.DisableConvertFilenameUnderscores {
				tablePath = path.Join(filepath.Dir(config.Filename), strings.ReplaceAll(tableName, "_", " ")+string(DBF))
			}
			tableConfig = &Config{
				Filename:                          tablePath,
				Converter:                         config.Converter,
				Exclusive:                         config.Exclusive,
				Untested:                          config.Untested,
				TrimSpaces:                        config.TrimSpaces,
				DisableConvertFilenameUnderscores: config.DisableConvertFilenameUnderscores,
				ReadOnly:                          config.ReadOnly,
				WriteLock:                         config.WriteLock,
				ValidateCodePage:                  config.ValidateCodePage,
				InterpretCodePage:                 config.InterpretCodePage,
			}
		}
		// Load the table
		table, err := OpenTable(tableConfig)
		if err != nil {
			return nil, WrapError(err)
		}
		if table != nil {
			tables[tableName] = table
		}
	}
	return &Database{file: databaseTable, tables: tables}, nil
}

// Close the database file and all related tables
func (db *Database) Close() error {
	for _, table := range db.tables {
		if err := table.Close(); err != nil {
			return WrapError(err)
		}
	}
	return db.file.Close()
}

// Returns all table of the database
func (db *Database) Tables() map[string]*File {
	return db.tables
}

// Returns the names of every table in the database
func (db *Database) Names() []string {
	names := make([]string, 0)
	for name := range db.tables {
		names = append(names, name)
	}
	return names
}

// Returns the complete database schema
func (db *Database) Schema() map[string][]*Column {
	schema := make(map[string][]*Column)
	for name, table := range db.tables {
		schema[name] = table.Columns()
	}
	return schema
}

// buildTableConfig creates a table config using the appropriate provider
func buildTableConfig(config *Config, tableName string) (*Config, error) {
	if config.TableProvider != nil {
		dbfData, memoData, err := config.TableProvider(tableName)
		if err != nil {
			return nil, NewErrorf("failed to get data for table %s: %v", tableName, err)
		}
		if dbfData == nil {
			return nil, nil // Skip if no data provided
		}

		return &Config{
			Data:                              dbfData,
			MemoData:                          memoData,
			Converter:                         config.Converter,
			Untested:                          config.Untested,
			TrimSpaces:                        config.TrimSpaces,
			DisableConvertFilenameUnderscores: config.DisableConvertFilenameUnderscores,
			ReadOnly:                          config.ReadOnly,
			WriteLock:                         config.WriteLock,
			ValidateCodePage:                  config.ValidateCodePage,
			InterpretCodePage:                 config.InterpretCodePage,
		}, nil
	}

	if config.TableReaderProvider != nil {
		dbfReader, memoReader, err := config.TableReaderProvider(tableName)
		if err != nil {
			return nil, NewErrorf("failed to get readers for table %s: %v", tableName, err)
		}
		if dbfReader == nil {
			return nil, nil // Skip if no reader provided
		}

		return &Config{
			Reader:                            dbfReader,
			MemoReader:                        memoReader,
			Converter:                         config.Converter,
			Untested:                          config.Untested,
			TrimSpaces:                        config.TrimSpaces,
			DisableConvertFilenameUnderscores: config.DisableConvertFilenameUnderscores,
			ReadOnly:                          config.ReadOnly,
			WriteLock:                         config.WriteLock,
			ValidateCodePage:                  config.ValidateCodePage,
			InterpretCodePage:                 config.InterpretCodePage,
		}, nil
	}

	return nil, NewError("when using Data or Reader for database, you must provide TableProvider or TableReaderProvider")
}
