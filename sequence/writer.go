// Package sequence provides Writer.
//
//spellchecker:words sequence
package sequence

//spellchecker:words sync atomic
import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Writer wraps an io.Writer and allows sequential writes.
//
// Writes are passed to the original writes as long as no errors occur.
// If a write fails a [PreviousWriteFailedError] is returned instead.
//
// A Writer keeps track of overall number of bytes written and the previous error.
// It can be retrieved using the [Sum] method, and reset using [Reset].
//
// The [Write] and [WriteString] methods may be used concurrently by multiple goroutines.
// The writes to the underlying writer are serialized automatically.
//
// Concurrent writes of the Writer field or other methods are not safe.
type Writer struct {
	hadError atomic.Bool // do
	l        sync.Mutex  // lock (if needed)

	n   int
	err error

	Writer io.Writer
}

// write performs the write operation w
//
// - if an error occurred previously, w is not called and 0, a [PreviousWriteFailed] is returned
// - if no error occurred, w is called and the state is updated
func (sw *Writer) write(w func() (int, error)) (int, error) {
	// fast path: we had an error, just return it!
	if sw.hadError.Load() {
		return 0, PreviousWriteFailedError{err: sw.err}
	}

	// slow path
	sw.l.Lock()
	defer sw.l.Unlock()

	// if there was an error, return it and don't do a write
	if sw.err != nil {
		return 0, PreviousWriteFailedError{err: sw.err}
	}

	// call the writer
	n, err := w()

	// update the state
	sw.n += n
	sw.err = err

	// if we had an error don't need to lock anymore!
	if err != nil {
		sw.hadError.Store(true)
	}

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
	sw.hadError.Store(false)
}

// PreviousWriteFailedError indicates that a previous write returned an error.
// The previous error can be retrieved via the Unwrap() method.
type PreviousWriteFailedError struct {
	err error
}

func (pw PreviousWriteFailedError) Error() string {
	return fmt.Sprintf("previous write failed: %s", pw.err)
}

func (pw PreviousWriteFailedError) Unwrap() error {
	return pw.err
}
