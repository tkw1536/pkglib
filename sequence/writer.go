// Package sequence provides Writer.
//
//spellchecker:words sequence
package sequence

import (
	"fmt"
	"io"
)

// Writer wraps an io.Writer.
//
// The first write behaves like a normal write would.
// Subsequent writes are only passed through as long as no errors occur.
// In case a write is not passed through, a [PreviousWriteFailed] is returned instead.
//
// A Writer keeps track of overall number of bytes written and the previous error.
// It can be retrieved using the [Sum] method, and reset using [Reset].
//
// The zero value is ready to use once the Writer field is set.
// A Writer is not safe to be used concurrently.
type Writer struct {
	n   int
	err error

	Writer io.Writer
}

// write performs the write operation w
//
// - if an error occurred previously, w is not called and 0, PreviousWriteFailed is returned
// - if no error occurred, w is called and the state is updated
func (sw *Writer) write(w func() (int, error)) (int, error) {
	// if there was an error, return it and don't do a write
	if sw.err != nil {
		return 0, PreviousWriteFailed{err: sw.err}
	}

	// call the writer
	n, err := w()

	// update the state
	sw.n += n
	sw.err = err

	// and return
	return n, err
}

// Write writes p to this Writer as described in the struct description.
func (sw *Writer) Write(p []byte) (int, error) {
	return sw.write(func() (int, error) {
		return sw.Writer.Write(p)
	})
}

// WriteString writes s to this SequenceWriter as described in the struct description.
func (sw *Writer) WriteString(s string) (int, error) {
	return sw.write(func() (int, error) {
		return io.WriteString(sw.Writer, s)
	})
}

// Sum returns the total number of bytes written the underlying writer, and any error that occurred.
// The underlying error is not wrapped.
func (cw *Writer) Sum() (int, error) {
	return cw.n, cw.err
}

// Reset resets the underlying error and total count of bytes written.
// Future writes will again be passed through to the underlying writer.
func (sw *Writer) Reset() {
	sw.err = nil
	sw.n = 0
}

// PreviousWriteFailed indicates that a previous write returned an error.
// The previous error can be retrieved via the Unwrap() method.
type PreviousWriteFailed struct {
	err error
}

func (pw PreviousWriteFailed) Error() string {
	return fmt.Sprintf("previous write failed: %s", pw.err)
}

func (pw PreviousWriteFailed) Unwrap() error {
	return pw.err
}
