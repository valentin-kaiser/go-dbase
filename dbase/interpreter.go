package dbase

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Converts raw column data to the correct type for the given column
// For C and M columns a charset conversion is done
// For M columns the data is read from the memo file
// At this moment not all FoxPro column types are supported.
// When reading column values, the value returned by this package is always `interface{}`.
//
// The supported column types with their return Go types are:
//
// | Column Type | Column Type Name | Golang type |
// | ----------- | ---------------- | ----------- |
// | B | Double | float64 |
// | C | Character | string |
// | D | Date | time.Time |
// | F | Float | float64 |
// | I | Integer | int32 |
// | L | Logical | bool |
// | M | Memo | string |
// | M | Memo (Binary) | []byte |
// | N | Numeric (0 decimals) | int64 |
// | N | Numeric (with decimals) | float64 |
// | T | DateTime | time.Time |
// | Y | Currency | float64 |
//
// Not all available column types have been implemented because we don't use them in our DBFs
func (file *File) Interpret(raw []byte, column *Column) (interface{}, error) {
	if len(raw) != int(column.Length) {
		return nil, NewErrorf("invalid length %v Bytes != %v Bytes at column field: %v", len(raw), column.Length, column.Name())
	}

	// A switch is used instead of a lookup table because Interpret runs once per
	// column of every row. Building a map of bound method values here would
	// allocate the map and all its closures on every single call.
	switch DataType(column.DataType) {
	// M values contain the address in the FPT file from where to read data
	case Memo:
		return file.parseMemo(raw, column)
	// C values are stored as strings, the returned string is not trimmed
	case Character:
		return file.parseCharacter(raw, column)
	// I values are stored as numeric values
	case Integer:
		return file.parseInteger(raw, column)
	// Y values are currency values stored as ints with 4 decimal places
	case Currency:
		return file.parseCurrency(raw, column)
	// F values are stored as string values
	case Float:
		return file.parseFloat(raw, column)
	// B (double) values are stored as numeric values
	case Double:
		return file.parseDouble(raw, column)
	// D values are stored as string in format YYYYMMDD, convert to time.Time
	case Date:
		return file.parseDate(raw, column)
	// T values are stores as two 4 byte integers
	//  integer one is the date in julian format
	//  integer two is the number of milliseconds since midnight
	case DateTime:
		return file.parseDateTime(raw, column)
	// L values are stored as strings T or F, we only check for T, the rest is false...
	case Logical:
		return file.parseLogical(raw, column)
	// N values are stored as string values, if no decimals return as int64, if decimals treat as float64
	case Numeric:
		return file.parseNumeric(raw, column)
	// V and Q values just return the raw value
	case Varchar:
		return file.parseVarchar(raw, column)
	case Varbinary:
		return file.parseVarbinary(raw, column)
	// W, P and G values just return the raw value
	case Blob, Picture, General:
		return file.parseRaw(raw, column)
	default:
		return nil, NewErrorf("unsupported column data type: %s at column field: %v", DataType(column.DataType), column.Name())
	}
}

// Represent converts column data to the byte representation of the columns data type
// For M values the data is written to the memo file and the address is returned
func (file *File) Represent(field *Field, padding bool) ([]byte, error) {
	if field.GetValue() == nil {
		return make([]byte, field.column.Length), nil
	}

	// See Interpret: a switch avoids allocating a map of bound method values on
	// every call.
	switch DataType(field.column.DataType) {
	// M values contain the address in the FPT file from where to read data
	case Memo:
		return file.getMemoRepresentation(field, padding)
	// C values are stored as strings, the returned string is not trimmed
	case Character:
		return file.getCharacterRepresentation(field, padding)
	// I values (int32)
	case Integer:
		return file.getIntegerRepresentation(field, padding)
	// Y (currency)
	case Currency:
		return file.getCurrencyRepresentation(field, padding)
	// F (Float)
	case Float:
		return file.getFloatRepresentation(field, padding)
	// B (double)
	case Double:
		return file.getDoubleRepresentation(field, padding)
	// D values are stored as string in format YYYYMMDD, convert to time.Time
	case Date:
		return file.getDateRepresentation(field, padding)
	// T values are stores as two 4 byte integers
	//  integer one is the date in julian format
	//  integer two is the number of milliseconds since midnight
	case DateTime:
		return file.getDateTimeRepresentation(field, padding)
	// L (bool) values are stored as strings T or F, we only check for T, the rest is false...
	case Logical:
		return file.getLogicalRepresentation(field, padding)
	// N values are stored as string values, if no decimals return as int64, if decimals treat as float64
	case Numeric:
		return file.getNumericRepresentation(field, padding)
	// V and Q values just return the raw value
	case Varchar:
		return file.getVarcharRepresentation(field, padding)
	case Varbinary:
		return file.getVarbinaryRepresentation(field, padding)
	// W, P and G values just return the raw value
	case Blob, Picture, General:
		return file.getRawRepresentation(field, padding)
	default:
		return nil, NewErrorf("unsupported column data type: %s at column field: %v", DataType(field.column.DataType), field.Name())
	}
}

// Returns the value from the memo file as string or []byte
func (file *File) parseMemo(raw []byte, column *Column) (interface{}, error) {
	// M values contain the address in the FPT file from where to read data
	if isEmptyBytes(raw) {
		return []byte{}, nil
	}
	memo, isText, err := file.ReadMemo(raw, column)
	if err != nil {
		return nil, NewErrorf("parsing memo failed at column field: %v failed", column.Name()).Details(err)
	}
	if isText {
		return string(memo), nil
	}
	return memo, nil
}

// Saves the value to the memo file and returns the address in the FPT file
func (file *File) getMemoRepresentation(field *Field, _ bool) ([]byte, error) {
	memo := make([]byte, 0)
	txt := false
	s, sok := field.value.(string)
	if sok {
		var err error
		memo, err = fromUtf8String([]byte(s), file.config.Converter)
		if err != nil {
			return nil, NewErrorf("parsing from utf8 string at column field: %v failed", field.Name()).Details(err)
		}
		txt = true
	}
	m, ok := field.value.([]byte)
	if ok {
		var err error
		memo, err = fromUtf8String(m, file.config.Converter)
		if err != nil {
			return nil, NewErrorf("parsing from utf8 string at column field: %v failed", field.Name()).Details(err)
		}
		txt = false
	}
	if !ok && !sok {
		return nil, NewErrorf("invalid type for memo field: %T", field.value)
	}
	address, err := file.WriteMemo(field.memoPos, memo, txt, len(memo))
	if err != nil {
		return nil, WrapError(err)
	}
	return address, nil
}

// Returns the value as string
func (file *File) parseCharacter(raw []byte, column *Column) (interface{}, error) {
	if len(raw) == 0 {
		return "", nil
	}
	if len(raw) > MaxCharacterLength {
		return NewErrorf("invalid length %v bytes > %v bytes at column field: %v", len(raw), MaxCharacterLength, column.Name()), nil
	}
	// C values are stored as strings, the returned string is not trimmed
	str, err := toUTF8String(raw, file.config.Converter)
	if err != nil {
		return str, NewErrorf("parsing to utf8 string failed at column field: %v failed", column.Name()).Details(err)
	}
	return str, nil
}

// Returns the string value as byte representation
func (file *File) getCharacterRepresentation(field *Field, skipSpacing bool) ([]byte, error) {
	// C values are stored as strings, the returned string is not trimmed
	c, ok := field.value.(string)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected string on column field: %v", field.value, field.Name())
	}
	raw := make([]byte, field.column.Length)
	bin, err := fromUtf8String([]byte(c), file.config.Converter)
	if err != nil {
		return nil, NewErrorf("parsing from utf8 string at column field: %v failed", field.Name()).Details(err)
	}
	if len(bin) > MaxCharacterLength {
		return nil, NewErrorf("invalid length %v bytes > %v bytes at column field: %v", len(bin), MaxCharacterLength, field.Name())
	}
	if skipSpacing {
		return bin, nil
	}
	bin = appendSpaces(bin, int(field.column.Length))
	copy(raw, bin)
	if len(raw) > int(field.column.Length) {
		return nil, NewErrorf("invalid length %v bytes > %v bytes at column field: %v", len(raw), field.column.Length, field.Name())
	}
	return raw, nil
}

// Returns the value as int32
func (file *File) parseInteger(raw []byte, _ *Column) (interface{}, error) {
	return int32(binary.LittleEndian.Uint32(raw)), nil
}

// Returns the int32 value as byte representation
func (file *File) getIntegerRepresentation(field *Field, _ bool) ([]byte, error) {
	// I values (int32)
	i, ok := field.value.(int32)
	if !ok {
		f, ok := field.value.(float64)
		if !ok {
			return nil, NewErrorf("invalid data type %T, expected int32 at column field: %v", field.value, field.Name())
		}
		// check for lower and uppper bounds to prevent overflow
		if f > 0 && f <= math.MaxInt32 {
			i = int32(f)
		}
	}
	raw := make([]byte, field.column.Length)
	bin, err := toBinary(i)
	if err != nil {
		return nil, NewErrorf("converting to binary at column field: %v failed", field.Name()).Details(err)
	}
	copy(raw, bin)
	if len(raw) != int(field.column.Length) {
		return nil, NewErrorf("invalid length %v bytes != %v bytes at column field: %v", len(raw), field.column.Length, field.Name())
	}
	return raw, nil
}

// Returns the value as float64
func (file *File) parseCurrency(raw []byte, _ *Column) (interface{}, error) {
	return float64(int64(binary.LittleEndian.Uint64(raw))) / 10000, nil
}

// Returns the float64 value as byte representation
func (file *File) getCurrencyRepresentation(field *Field, _ bool) ([]byte, error) {
	f, ok := field.value.(float64)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected float64 at column field: %v", field.value, field.Name())
	}
	// Cast to int64 and multiply by 10000 to get the value as int64 with 4 decimals
	i := int64(math.Round(f * 10000))
	raw := make([]byte, field.column.Length)
	bin, err := toBinary(i)
	if err != nil {
		return nil, NewErrorf("converting to binary at column field: %v failed", field.Name()).Details(err)
	}
	copy(raw, bin)
	if len(raw) != int(field.column.Length) {
		return nil, NewErrorf("invalid length %v bytes != %v bytes at column field: %v", len(raw), field.column.Length, field.Name())
	}
	return raw, nil
}

// Returns the value as float64
func (file *File) parseFloat(raw []byte, column *Column) (interface{}, error) {
	f, err := parseFloat(raw)
	if err != nil {
		return f, NewErrorf("parsing float at column field: %v failed", column.Name()).Details(err)
	}
	return f, nil
}

// Returns the float64 value as byte representation
func (file *File) getFloatRepresentation(field *Field, skipSpacing bool) ([]byte, error) {
	b, ok := field.value.(float64)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected float64 at column field: %v", field.value, field.Name())
	}
	var bin []byte
	if b == float64(int64(b)) {
		// if the value has no decimals, store as integer
		bin = []byte(strconv.FormatInt(int64(b), 10))
	} else {
		// if the value is a float, store as float
		expression := fmt.Sprintf("%%.%df", field.column.Decimals)
		bin = []byte(fmt.Sprintf(expression, field.value))
	}
	if skipSpacing {
		return bin, nil
	}
	return prependSpaces(bin, int(field.column.Length)), nil
}

// Returns the value as float64
func (file *File) parseDouble(raw []byte, _ *Column) (interface{}, error) {
	return math.Float64frombits(binary.LittleEndian.Uint64(raw)), nil
}

// Returns the float64 value as byte representation
func (file *File) getDoubleRepresentation(field *Field, _ bool) ([]byte, error) {
	b, ok := field.value.(float64)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected float64 at column field: %v", field.value, field.Name())
	}
	raw := make([]byte, field.column.Length)
	bin, err := toBinary(b)
	if err != nil {
		return nil, NewErrorf("converting to binary at column field: %v failed", field.Name()).Details(err)
	}
	copy(raw, bin)
	if len(raw) != int(field.column.Length) {
		return nil, NewErrorf("invalid length %v bytes != %v bytes at column field: %v", len(raw), field.column.Length, field.Name())
	}
	return raw, nil
}

// Returns the value as time.Time
func (file *File) parseDate(raw []byte, column *Column) (interface{}, error) {
	// D values are stored as string in format YYYYMMDD, convert to time.Time
	date, err := parseDate(raw)
	if err != nil {
		return date, NewErrorf("parsing to date at column field: %v failed", column.Name()).Details(err)
	}
	return date, nil
}

// Get the time.Time value as byte representation
func (file *File) getDateRepresentation(field *Field, _ bool) ([]byte, error) {
	d, ok := field.value.(time.Time)
	if !ok {
		s, ok := field.value.(string)
		if !ok {
			return nil, NewErrorf("invalid data type %T, expected time.Time at column field: %v", field.value, field.Name())
		}

		t := time.Time{}
		var err error
		if len(s) > 0 {
			t, err = time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, NewErrorf("parsing time failed at column field: %v failed", field.Name()).Details(err)
			}
		}

		d = t
	}

	raw := make([]byte, field.column.Length)
	bin := []byte(strings.Repeat(" ", int(field.column.Length)))
	if !d.IsZero() {
		bin = []byte(d.Format("20060102"))
	}
	copy(raw, bin)
	if len(raw) != int(field.column.Length) {
		return nil, NewErrorf("invalid length %v bytes != %v bytes at column field: %v", len(raw), field.column.Length, field.Name())
	}
	return raw, nil
}

// Returns the value as time.Time
func (file *File) parseDateTime(raw []byte, _ *Column) (interface{}, error) {
	return parseDateTime(raw), nil
}

// Get the time.Time value as byte representation consisting of 4 bytes for julian date and 4 bytes for time
func (file *File) getDateTimeRepresentation(field *Field, _ bool) ([]byte, error) {
	t, ok := field.value.(time.Time)
	if !ok {
		s, ok := field.value.(string)
		if !ok {
			return nil, NewErrorf("invalid data type %T, expected time.Time at column field: %v", field.value, field.Name())
		}
		parsedTime, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, NewErrorf("parsing time failed at column field: %v failed", field.Name()).Details(err)
		}
		t = parsedTime
	}
	if t.IsZero() {
		return make([]byte, field.column.Length), nil
	}
	raw := make([]byte, 8)
	i := julianDate(t.Year(), int(t.Month()), t.Day())
	date, err := toBinary(uint64(i))
	if err != nil {
		return nil, NewErrorf("time conversion at column field: %v failed", field.Name()).Details(err)
	}
	copy(raw[:4], date)
	millis := t.Hour()*3600000 + t.Minute()*60000 + t.Second()*1000 + t.Nanosecond()/1000000
	time, err := toBinary(uint64(millis))
	if err != nil {
		return nil, NewErrorf("binary conversion at column field: %v failed", field.Name()).Details(err)
	}
	copy(raw[4:], time)
	if len(raw) != int(field.column.Length) {
		return nil, NewErrorf("invalid length %v bytes != %v bytes at column field: %v", len(raw), field.column.Length, field.Name())
	}
	return raw, nil
}

// Return the value (T or F) as bool
func (file *File) parseLogical(raw []byte, _ *Column) (interface{}, error) {
	return string(raw) == "T", nil
}

// Get the bool value as byte representation (T or F)
func (file *File) getLogicalRepresentation(field *Field, _ bool) ([]byte, error) {
	l, ok := field.value.(bool)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected bool at column field: %v", field.value, field.Name())
	}
	raw := []byte("F")
	if l {
		return []byte("T"), nil
	}
	return raw, nil
}

// Get the raw value as byte representation
func (file *File) parseRaw(raw []byte, _ *Column) (interface{}, error) {
	return raw, nil
}

// Get the raw value as byte representation (only type check for []byte is performed)
func (file *File) getRawRepresentation(field *Field, _ bool) ([]byte, error) {
	// If string is passed, convert to []byte
	if s, ok := field.value.(string); ok {
		return []byte(s), nil
	}
	raw, ok := field.value.([]byte)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected []byte at column field: %v", field.value, field.Name())
	}
	return raw, nil
}

// Returns the value as integer or float64
func (file *File) parseNumeric(raw []byte, column *Column) (interface{}, error) {
	if column.Decimals == 0 {
		i, err := parseNumericInt(raw)
		if err != nil {
			return i, NewErrorf("parsing numeric int at column field: %v failed", column.Name()).Details(err)
		}
		return i, nil
	}

	return file.parseFloat(raw, column)
}

// Get the integer or float value as byte representation (supports int32/int64, uint/uint32/uint64, float32/float64).
func (file *File) getNumericRepresentation(field *Field, skipSpacing bool) ([]byte, error) {
	var bin []byte

	switch v := field.value.(type) {
	case float64:
		// If no decimals, store as integer; otherwise format as float with Decimals precision
		if v == math.Trunc(v) {
			bin = []byte(strconv.FormatInt(int64(v), 10))
		} else {
			bin = []byte(strconv.FormatFloat(v, 'f', int(field.column.Decimals), 64))
		}

	case float32:
		f := float64(v)
		if f == math.Trunc(f) {
			bin = []byte(strconv.FormatInt(int64(f), 10))
		} else {
			bin = []byte(strconv.FormatFloat(f, 'f', int(field.column.Decimals), 64))
		}

	case int64:
		bin = []byte(strconv.FormatInt(v, 10))

	case int32:
		bin = []byte(strconv.FormatInt(int64(v), 10))

	case int16:
		bin = []byte(strconv.FormatInt(int64(v), 10))

	case int8:
		bin = []byte(strconv.FormatInt(int64(v), 10))

	case int:
		bin = []byte(strconv.FormatInt(int64(v), 10))

	case uint64:
		bin = []byte(strconv.FormatUint(v, 10))

	case uint32:
		bin = []byte(strconv.FormatUint(uint64(v), 10))

	case uint16:
		bin = []byte(strconv.FormatUint(uint64(v), 10))

	case uint8:
		bin = []byte(strconv.FormatUint(uint64(v), 10))

	case uint:
		bin = []byte(strconv.FormatUint(uint64(v), 10))

	case *big.Int:
		bin = []byte(v.String())

	default:
		return nil, NewErrorf(
			"invalid data type %T, expected int32, int64, uint, uint32, uint64, float32 or float64 at column field: %v",
			field.value, field.Name(),
		)
	}

	if skipSpacing {
		return bin, nil
	}
	return prependSpaces(bin, int(field.column.Length)), nil
}

func (file *File) parseVarchar(raw []byte, column *Column) (interface{}, error) {
	varlen, null, err := file.ReadNullFlag(uint64(file.table.rowPointer), column)
	if err != nil {
		return nil, NewErrorf("reading null flag at column field: %v failed", column.Name()).Details(err)
	}
	if null {
		return []byte{}, nil
	}
	if varlen {
		length := raw[len(raw)-1]
		raw = raw[:length]
	}
	return string(raw), nil
}

func (file *File) getVarcharRepresentation(field *Field, _ bool) ([]byte, error) {
	s, ok := field.value.(string)
	if ok {
		return []byte(s), nil
	}
	m, ok := field.value.([]byte)
	if ok {
		return m, nil
	}
	return nil, NewErrorf("invalid data type %T, expected string at column field: %v", field.value, field.Name())
}

func (file *File) parseVarbinary(raw []byte, column *Column) (interface{}, error) {
	varlen, null, err := file.ReadNullFlag(uint64(file.table.rowPointer), column)
	if err != nil {
		return nil, NewErrorf("reading null flag at column field: %v failed", column.Name()).Details(err)
	}
	if null {
		return []byte{}, nil
	}
	if varlen {
		length := raw[len(raw)-1]
		raw = raw[:length]
	}
	return raw, nil
}

func (file *File) getVarbinaryRepresentation(field *Field, _ bool) ([]byte, error) {
	raw, ok := field.value.([]byte)
	if !ok {
		return nil, NewErrorf("invalid data type %T, expected []byte at column field: %v", field.value, field.Name())
	}
	return raw, nil
}

func isEmptyBytes(b []byte) bool {
	if len(b) == 0 {
		return true
	}

	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
