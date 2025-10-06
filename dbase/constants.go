package dbase

import (
	"math"
)

// Supported and testet file versions - other files may work but are not tested
// The file version check has to be bypassed when opening a file type that is not supported
// https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/st4a0s68(v=vs.71)
type FileVersion byte

// Supported and testet file types - other file types may work but are not tested
const (
	// FoxPro represents FoxPro file format (0x30)
	FoxPro FileVersion = 0x30
	// FoxProAutoincrement represents FoxPro file format with autoincrement support (0x31)
	FoxProAutoincrement FileVersion = 0x31
	// FoxProVar represents FoxPro file format with variable length fields (0x32)
	FoxProVar FileVersion = 0x32
)

// Not tested file versions - these may work but are not officially supported
const (
	// FoxBase represents FoxBase file format (0x02)
	FoxBase FileVersion = 0x02
	// FoxBase2 represents FoxBase 2 file format (0xFB)
	FoxBase2 FileVersion = 0xFB
	// FoxBasePlus represents FoxBase Plus file format (0x03)
	FoxBasePlus FileVersion = 0x03
	// DBaseSQLTable represents dBase SQL table format (0x43)
	DBaseSQLTable FileVersion = 0x43
	// FoxBasePlusMemo represents FoxBase Plus with memo support (0x83)
	FoxBasePlusMemo FileVersion = 0x83
	// DBaseMemo represents dBase with memo support (0x8B)
	DBaseMemo FileVersion = 0x8B
	// DBaseSQLMemo represents dBase SQL with memo support (0xCB)
	DBaseSQLMemo FileVersion = 0xCB
	// FoxPro2Memo represents FoxPro 2 with memo support (0xF5)
	FoxPro2Memo FileVersion = 0xF5
)

// Allowed file extensions for the different file types
type FileExtension string

const (
	DBC FileExtension = ".DBC" // Database file extension
	DCT FileExtension = ".DCT" // Database container file extension
	DBF FileExtension = ".DBF" // Table file extension
	FPT FileExtension = ".FPT" // Memo file extension
	SCX FileExtension = ".SCX" // Form file extension
	LBX FileExtension = ".LBX" // Label file extension
	MNX FileExtension = ".MNX" // Menu file extension
	PJX FileExtension = ".PJX" // Project file extension
	RPX FileExtension = ".RPX" // Report file extension
	VCX FileExtension = ".VCX" // Visual class library file extension
)

// Important byte markers for the dBase file
type Marker byte

const (
	// Null represents a null byte marker (0x00)
	Null Marker = 0x00
	// Blank represents a space/blank marker (0x20)
	Blank Marker = 0x20
	// ColumnEnd represents the end of column definitions marker (0x0D)
	ColumnEnd Marker = 0x0D
	// Active represents an active (non-deleted) row marker, same as Blank
	Active Marker = Blank
	// Deleted represents a deleted row marker (0x2A)
	Deleted Marker = 0x2A
	// EOFMarker represents the end of file marker (0x1A)
	EOFMarker Marker = 0x1A
)

// Table flags indicate the type of the table
// https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/st4a0s68(v=vs.71)
type TableFlag byte

const (
	// StructuralFlag indicates the table is a structural table (0x01)
	StructuralFlag TableFlag = 0x01
	// MemoFlag indicates the table has an associated memo file (0x02)
	MemoFlag TableFlag = 0x02
	// DatabaseFlag indicates the table is part of a database (0x04)
	DatabaseFlag TableFlag = 0x04
)

// Defined checks if the table flag is set to the given flag value.
func (t TableFlag) Defined(flag byte) bool {
	return t&TableFlag(flag) == t
}

// Flag represents a general purpose flag type for bit operations.
type Flag byte

// Has checks if the flag has the specified mask bits set.
func (f Flag) Has(mask byte) bool {
	return byte(f)&mask != 0
}

// HasAll checks if the flag has all the specified mask bits set.
func (f Flag) HasAll(mask byte) bool {
	return byte(f)&mask == mask
}

// Column flags indicate whether a column is hidden, can be null, is binary or is autoincremented
type ColumnFlag byte

const (
	// HiddenFlag indicates the column is hidden from view (0x01)
	HiddenFlag ColumnFlag = 0x01
	// NullableFlag indicates the column can contain null values (0x02)
	NullableFlag ColumnFlag = 0x02
	// BinaryFlag indicates the column contains binary data (NOCPTRANS) (0x04)
	BinaryFlag ColumnFlag = 0x04 // NOCPTRANS
	// AutoincrementFlag indicates the column is auto-incremented (0x08)
	AutoincrementFlag ColumnFlag = 0x08
)

// DataType defines the possible types of a column
type DataType byte

const (
	Character DataType = 0x43 // C - Character (string)
	Currency  DataType = 0x59 // Y - Currency (float64)
	Double    DataType = 0x42 // B - Double (float64)
	Date      DataType = 0x44 // D - Date (time.Time)
	DateTime  DataType = 0x54 // T - DateTime (time.Time)
	Float     DataType = 0x46 // F - Float (float64)
	Integer   DataType = 0x49 // I - Integer (int32)
	Logical   DataType = 0x4C // L - Logical (bool)
	Memo      DataType = 0x4D // M - Memo (string)
	Numeric   DataType = 0x4E // N - Numeric (int64)
	Blob      DataType = 0x57 // W - Blob ([]byte)
	General   DataType = 0x47 // G - General (string)
	Picture   DataType = 0x50 // P - Picture (string)
	Varbinary DataType = 0x51 // Q - Varbinary ([]byte)
	Varchar   DataType = 0x56 // V - Varchar (string)
)

// String returns the type of the column as string representation.
func (t DataType) String() string {
	return string(t)
}

// nullFlagColumn is a reserved column name that is placed at the end of the column list
// It indicates wether a column is nullable or has a variable length. The value of the column
// is a byte arry where one bit indicates wether the column is nullable and another bit indicates
// wether the column has a variable length.
var nullFlagColumn = [11]byte{0x5F, 0x4E, 0x75, 0x6C, 0x6C, 0x46, 0x6C, 0x61, 0x67, 0x73}

// dBase format limits and constraints
// https://learn.microsoft.com/en-us/previous-versions/visualstudio/foxpro/3kfd3hw9(v=vs.71)
const (
	// MaxColumnNameLength is the maximum length for column names (10 characters)
	MaxColumnNameLength = 10
	// MaxCharacterLength is the maximum length for character fields (254 characters)
	MaxCharacterLength = 254
	// MaxNumericLength is the maximum length for numeric fields (20 digits)
	MaxNumericLength = 20
	// MaxFloatLength is the maximum length for float fields (20 digits)
	MaxFloatLength = 20
	// MaxIntegerValue is the maximum value for integer fields
	MaxIntegerValue = math.MaxInt32
	// MinIntegerValue is the minimum value for integer fields
	MinIntegerValue = math.MinInt32
	// MaxFieldsPerRecord is the maximum number of fields per record (255)
	MaxFieldsPerRecord = 255
	// MaxCharactersPerRecord is the maximum number of characters per record (65500)
	MaxCharactersPerRecord = 65500
	// MaxTableFileSize is the maximum size of a table file in bytes (2GB)
	MaxTableFileSize = 2 << 30
	// MaxRecordsPerTable is the maximum number of records per table (1 billion)
	MaxRecordsPerTable = 1000000000
	// MaxIndexKeyLength is the maximum length for index keys (100 characters)
	MaxIndexKeyLength = 100
	// MaxCompactIndexKeyLength is the maximum length for compact index keys (240 characters)
	MaxCompactIndexKeyLength = 240
	// NumericPrecisionDigits is the number of precision digits for numeric calculations (16)
	NumericPrecisionDigits = 16
)
