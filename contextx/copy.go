package contextx

import (
	"context"
	"io"
	"os"
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
		// close the source file
		// and set a read deadline!
		if closer, ok := src.(io.Closer); ok {
			closer.Close()
		}
		if file, ok := src.(*os.File); ok {
			file.SetReadDeadline(time.Now())
		}

		// close the destination file
		// and set a write deadline!
		if closer, ok := dst.(io.Closer); ok {
			closer.Close()
		}
		if file, ok := dst.(*os.File); ok {
			file.SetWriteDeadline(time.Now())
		}
	})
	return written, err
}
