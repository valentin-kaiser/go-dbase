package dbase

import (
	"bytes"
	"fmt"
	"io"
)

// BytesReadWriteSeeker wraps a byte slice to implement io.ReadWriteSeeker.
// This allows byte data to be used with GenericIO for reading dBase files from memory.
type BytesReadWriteSeeker struct {
	data   []byte
	reader *bytes.Reader
	pos    int64
}

// NewBytesReadWriteSeeker creates a new BytesReadWriteSeeker from a byte slice.
// The data is copied to ensure the wrapper owns the data.
func NewBytesReadWriteSeeker(data []byte) *BytesReadWriteSeeker {
	if data == nil {
		return nil
	}

	// Create a copy of the data to ensure we own it
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	return &BytesReadWriteSeeker{
		data:   dataCopy,
		reader: bytes.NewReader(dataCopy),
		pos:    0,
	}
}

// Read implements io.Reader.
func (b *BytesReadWriteSeeker) Read(p []byte) (n int, err error) {
	if b.reader == nil {
		return 0, io.EOF
	}

	n, err = b.reader.Read(p)
	b.pos += int64(n)
	return n, err
}

// Write implements io.Writer.
// Note: For dBase file reading, write operations are typically not needed,
// but this implementation allows the interface to be satisfied.
func (b *BytesReadWriteSeeker) Write(p []byte) (n int, err error) {
	if b.pos < 0 || b.pos > int64(len(b.data)) {
		return 0, fmt.Errorf("invalid seek position: %d", b.pos)
	}

	// If writing beyond current data, extend the slice
	endPos := b.pos + int64(len(p))
	if endPos > int64(len(b.data)) {
		// Extend the data slice
		newData := make([]byte, endPos)
		copy(newData, b.data)
		b.data = newData
	}

	// Write data at current position
	n = copy(b.data[b.pos:], p)
	b.pos += int64(n)

	// Update the reader to reflect changes
	b.reader = bytes.NewReader(b.data)
	b.reader.Seek(b.pos, io.SeekStart)

	return n, nil
}

// Seek implements io.Seeker.
func (b *BytesReadWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	if b.reader == nil {
		return 0, fmt.Errorf("reader is nil")
	}

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = b.pos + offset
	case io.SeekEnd:
		newPos = int64(len(b.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence value: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative seek position: %d", newPos)
	}

	b.pos = newPos
	return b.reader.Seek(offset, whence)
}

// Close implements io.Closer (optional, for interface compatibility).
func (b *BytesReadWriteSeeker) Close() error {
	// No resources to close for in-memory data
	return nil
}

// Size returns the current size of the underlying data.
func (b *BytesReadWriteSeeker) Size() int64 {
	return int64(len(b.data))
}

// Data returns a copy of the underlying data.
func (b *BytesReadWriteSeeker) Data() []byte {
	result := make([]byte, len(b.data))
	copy(result, b.data)
	return result
}
