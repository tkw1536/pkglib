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
// Writes (using the Write* methods) are passed to the original writer as long as no errors occur.
// If a write fails a [PreviousWriteFailedError] is returned instead.
// It keeps track of overall number of bytes written and the previous error.
//
// The Write* methods return an integer and wrap the error using [PreviousWriteFailedError].
// The underlying error can also be retrieved directly sing the [Writer.Sum] method.
//
// The [Writer.Write] and [Writer.WriteString] methods may be used concurrently by multiple goroutines.
// The writes to the underlying writer are serialized automatically.
//
// Concurrent writes of the Writer field or other methods are not safe.
type Writer struct {
	hadError atomic.Bool

	l   sync.Mutex // protects n and error
	err PreviousWriteFailedError

	Writer io.Writer
}

// performs the write operation described by w.
func (sw *Writer) write(w func() (int, error)) (int, error) {
	if !sw.hadError.Load() {
		return sw.writeSlow(w)
	}

	return 0, sw.err
}

func (sw *Writer) writeSlow(w func() (int, error)) (int, error) {
	sw.l.Lock()
	defer sw.l.Unlock()

	if sw.err.Err != nil {
		return 0, sw.err
	}

	n, err := w()
	sw.err.N += n
	sw.err.Err = err

	if err != nil {
		sw.hadError.Store(true)
	}

	return n, err
}

// Write writes p to the underlying writer.
func (sw *Writer) Write(p []byte) (int, error) {
	return sw.write(func() (int, error) {
		return sw.Writer.Write(p)
	})
}

// WriteString writes s to the underlying writer.
func (sw *Writer) WriteString(s string) (int, error) {
	return sw.write(func() (int, error) {
		return io.WriteString(sw.Writer, s)
	})
}

// Sum returns the total number of bytes written the underlying writer, and any error that occurred.
func (cw *Writer) Sum() (int, error) {
	return cw.err.N, cw.err.Err
}

// Reset resets the underlying error and total count of bytes written.
// Future writes will again be passed through to the underlying writer.
func (sw *Writer) Reset() {
	sw.err.Err = nil
	sw.err.N = 0
	sw.hadError.Store(false)
}

// PreviousWriteFailedError indicates that a previous write operation to [Writer.Write] failed.
type PreviousWriteFailedError struct {
	N   int   // number of bytes written before the error occurred
	Err error // error returned by [io.Writer.Write]
}

func (pw PreviousWriteFailedError) Error() string {
	s := ""
	if pw.N != 1 {
		s = "s"
	}

	return fmt.Sprintf("previous write failed after %d byte%s: %s", pw.N, s, pw.Err)
}

func (pw PreviousWriteFailedError) Unwrap() error {
	return pw.Err
}
