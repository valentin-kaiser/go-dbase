package dbase

import "io"

// IO is the interface for working with dBase files.
// It provides methods for opening, reading, writing, and managing dBase database files and their memo files.
// Three implementations are available:
// - WindowsIO (for direct file access on Windows)
// - UnixIO (for direct file access on Unix-like systems)
// - GenericIO (for custom file access implementing io.ReadWriteSeeker)
type IO interface {
	OpenTable(config *Config) (*File, error)
	Close(file *File) error
	Create(file *File) error
	ReadHeader(file *File) error
	WriteHeader(file *File) error
	ReadColumns(file *File) ([]*Column, *Column, error)
	WriteColumns(file *File) error
	ReadMemoHeader(file *File) error
	WriteMemoHeader(file *File, size int) error
	ReadMemo(file *File, address []byte, column *Column) ([]byte, bool, error)
	WriteMemo(address []byte, file *File, raw []byte, text bool, length int) ([]byte, error)
	ReadNullFlag(file *File, position uint64, column *Column) (bool, bool, error)
	ReadRow(file *File, position uint32) ([]byte, error)
	WriteRow(file *File, row *Row) error
	Search(file *File, field *Field, exactMatch bool) ([]*Row, error)
	GoTo(file *File, row uint32) error
	Skip(file *File, offset int64)
	Deleted(file *File) (bool, error)
}

// OpenTable opens a dBase database file (and the memo file if needed).
// The config parameter is required to specify either:
//   - IO: custom IO implementation (takes priority if provided)
//   - Data: DBF file content as bytes (with optional MemoData for FPT content)
//   - Reader: DBF file content as io.ReadWriteSeeker (with optional MemoReader)
//   - Filename: path to DBF file on filesystem (fallback option)
//
// If no IO is provided, one will be created based on available data sources.
func OpenTable(config *Config) (*File, error) {
	if config == nil {
		return nil, NewError("missing dbase configuration")
	}

	// Validate that exactly one data source is provided
	if err := config.validateDataSources(); err != nil {
		return nil, err
	}

	// If custom IO is already provided, use it directly
	if config.IO != nil {
		return config.IO.OpenTable(config)
	}

	// No custom IO provided, so create one based on available data sources
	if config.Data != nil || config.Reader != nil {
		// Create GenericIO for byte/reader data
		var dbfHandle, memoHandle io.ReadWriteSeeker

		if config.Reader != nil {
			dbfHandle = config.Reader
			memoHandle = config.MemoReader
		}

		if config.Data != nil {
			dbfHandle = NewBytesReadWriteSeeker(config.Data)
			if config.MemoData != nil {
				memoHandle = NewBytesReadWriteSeeker(config.MemoData)
			}
		}

		// Create a copy of config with GenericIO
		configCopy := *config
		configCopy.IO = GenericIO{
			Handle:        dbfHandle,
			RelatedHandle: memoHandle,
		}

		return configCopy.IO.OpenTable(&configCopy)
	}

	// Fall back to filesystem access with DefaultIO
	if config.Filename == "" {
		return nil, NewError("missing filename, data, or reader in configuration")
	}

	config.IO = DefaultIO
	return config.IO.OpenTable(config)
}

// Close closes all file handlers for the dBase file and its associated memo file.
func (file *File) Close() error {
	return file.defaults().io.Close(file)
}

// Create creates a new dBase database file (and the memo file if needed).
func (file *File) Create() error {
	file.isNew = true
	return file.defaults().io.Create(file)
}

// ReadHeader reads the dBase file header from the file handle.
func (file *File) ReadHeader() error {
	return file.defaults().io.ReadHeader(file)
}

// WriteHeader writes the header to the dBase file.
func (file *File) WriteHeader() error {
	return file.defaults().io.WriteHeader(file)
}

// ReadColumns reads column definitions from the dBase file header, starting at position 32,
// until it finds the header row terminator END_OF_COLUMN (0x0D).
func (file *File) ReadColumns() ([]*Column, *Column, error) {
	return file.defaults().io.ReadColumns(file)
}

// WriteColumns writes the column definitions to the end of the header in the dBase file.
func (file *File) WriteColumns() error {
	return file.defaults().io.WriteColumns(file)
}

// ReadMemoHeader reads the memo file header from the given file handle.
func (file *File) ReadMemoHeader() error {
	return file.defaults().io.ReadMemoHeader(file)
}

// WriteMemoHeader writes the memo header to the memo file.
// The size parameter specifies the number of blocks the new memo data will occupy.
func (file *File) WriteMemoHeader(size int) error {
	return file.defaults().io.WriteMemoHeader(file, size)
}

// ReadRow reads the raw row data of one row at the specified row position.
func (file *File) ReadRow(position uint32) ([]byte, error) {
	return file.defaults().io.ReadRow(file, position)
}

// WriteRow writes the raw row data to the specified row position in the dBase file.
func (file *File) WriteRow(row *Row) error {
	return file.defaults().io.WriteRow(file, row)
}

// ReadMemo reads one or more blocks from the memo file for the specified memo column.
// Returns the raw data and a boolean indicating if the data is text (true) or binary (false).
func (file *File) ReadMemo(address []byte, column *Column) ([]byte, bool, error) {
	return file.defaults().io.ReadMemo(file, address, column)
}

// WriteMemo writes memo data to the memo file and returns the address of the memo.
// The text parameter indicates whether the data is text (true) or binary (false).
// The length parameter specifies the length of the data to write.
func (file *File) WriteMemo(address []byte, data []byte, text bool, length int) ([]byte, error) {
	return file.defaults().io.WriteMemo(address, file, data, text, length)
}

// ReadNullFlag reads the null flag field at the end of the row.
// The null flag field indicates if the field has a variable length.
// Returns true as the first value if the field is variable length, and true as the second value if the field is null.
func (file *File) ReadNullFlag(position uint64, column *Column) (bool, bool, error) {
	return file.defaults().io.ReadNullFlag(file, position, column)
}

// Search searches for rows that contain the specified value in the given field.
// If exactMatch is true, only exact matches are returned; otherwise, partial matches are included.
func (file *File) Search(field *Field, exactMatch bool) ([]*Row, error) {
	return file.defaults().io.Search(file, field, exactMatch)
}

// GoTo sets the internal row pointer to the specified row number.
// Returns an EOF error if positioning beyond the end of file and positions the pointer at lastRow+1.
func (file *File) GoTo(row uint32) error {
	return file.defaults().io.GoTo(file, row)
}

// Skip adds the specified offset to the internal row pointer.
// If the result would position beyond the end of file, positions the pointer at lastRow+1.
// If the result would be negative, positions the pointer at 0.
// Note: This method does not skip deleted rows automatically.
func (file *File) Skip(offset int64) {
	file.defaults().io.Skip(file, offset)
}

// Deleted returns true if the row at the current internal row pointer position is marked as deleted.
func (file *File) Deleted() (bool, error) {
	return file.defaults().io.Deleted(file)
}

// GetIO returns the IO implementation currently being used by this file.
func (file *File) GetIO() IO {
	return file.io
}

// GetHandle returns the file handles being used (dBase file handle, memo file handle).
func (file *File) GetHandle() (interface{}, interface{}) {
	return file.handle, file.relatedHandle
}

// Sets the default if no io is set
func (file *File) defaults() *File {
	if file.io == nil {
		file.io = DefaultIO
	}
	return file
}

// ValidateFileVersion checks if the dBase file version is supported and tested.
// If untested is true, validation is bypassed and any version is accepted.
func ValidateFileVersion(version byte, untested bool) error {
	if untested {
		return nil
	}
	debugf("Validating file version: %d", version)
	switch version {
	default:
		return NewErrorf("untested DBF file version: %d (0x%x)", version, version)
	case byte(FoxPro), byte(FoxProAutoincrement), byte(FoxProVar):
		return nil
	}
}
