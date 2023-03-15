// Package stream provides NullStream
package stream

import (
	"io"
)

// Null is an io.ReadWriteCloser.
//
// Reads from it return 0 bytes and io.EOF.
// Writes and Closes succeed without doing anything.
//
// See also io.Discard.
var Null io.ReadWriteCloser = nullStream{}

// IsNullWriter checks if a writer is known to be a writer that discards any input
func IsNullWriter(writer io.Writer) bool {
	return writer == Null || writer == io.Discard
}

type nullStream struct{}

func (nullStream) Read(bytes []byte) (int, error) {
	return 0, io.EOF
}
func (nullStream) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Discard.(io.ReaderFrom).ReadFrom(r)
}

func (nullStream) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}
func (nullStream) WriteString(s string) (int, error) {
	return len(s), nil
}
func (nullStream) Close() error {
	return nil
}
