package contextx

import (
	"context"
	"io"
	"time"
)

// Copy copies from src to dst, cancelling the operation is cancelled.
// See io.Copy() for a description of the copy behaviour.
//
// The operation is cancelled by closing the src and destionation (if they support the Close() interface).
// Futhermore appropriate read and write deadlines are set.
// Either of these calls may not have any effect, depending on the underlying operation.
func Copy(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	// NOTE(twiesing): This function is not tested
	// Because there is no good way of testing if cancellation works!

	written, err, _ = Run2(ctx, func(start func()) (int64, error) {
		start()
		return io.Copy(dst, src)
	}, func() {
		CancelRead(src)
		CancelWrite(dst)
	})
	return written, err
}

// CancelRead attempts to cancel any in-progress and future reads on the given reader.
// In particular, this function sets the read deadline to the current time and closes the reader.
func CancelRead(reader io.Reader) {
	if srd, ok := reader.(interface{ SetReadDeadline(time.Time) error }); ok {
		srd.SetReadDeadline(time.Now())
	}

	if closer, ok := reader.(io.Closer); ok {
		closer.Close()
	}
}

// CancelWrite attempts to cancel any in-progress and future writes on the given writer.
// In particular, this function sets the write deadline to the current time and closes the writer.
func CancelWrite(writer io.Writer) {
	if swd, ok := writer.(interface{ SetWriteDeadline(time.Time) error }); ok {
		swd.SetWriteDeadline(time.Now())
	}

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}
