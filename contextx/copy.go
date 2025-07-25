//spellchecker:words contextx
package contextx

//spellchecker:words context errors time github pkglib errorsx
import (
	"context"
	"errors"
	"io"
	"time"

	"go.tkw01536.de/pkglib/errorsx"
)

// Copy copies from src to dst, stopping once ctx is closed.
// See [io.Copy] for a description of the copy behavior.
//
// The operation is cancelled by closing the src and destination, if they support the [io.Closer] interface.
// Furthermore appropriate read and write deadlines are set.
// Either of these calls may not have any effect, depending on the underlying operation.
func Copy(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	// NOTE: This function is not tested
	// Because there is no good way of testing if cancellation works!

	written, err, _ = Run2(ctx, func(start func()) (int64, error) {
		start()
		return io.Copy(dst, src)
	}, func() {
		// ignore any errors trying to cancel
		_ = cancelRead(src)
		_ = cancelWrite(dst)
	})
	return written, err
}

// cancelRead attempts to cancel any in-progress and future reads on the given reader.
// In particular, this function sets the read deadline to the current time and closes the reader.
func cancelRead(reader io.Reader) error {
	var errDeadline error
	var errCloser error

	if srd, ok := reader.(interface{ SetReadDeadline(time.Time) error }); ok {
		errDeadline = srd.SetReadDeadline(time.Now())
	}

	if closer, ok := reader.(io.Closer); ok {
		errCloser = closer.Close()
	}

	return errorsx.Combine(errDeadline, errCloser)
}

// cancelWrite attempts to cancel any in-progress and future writes on the given writer.
// In particular, this function sets the write deadline to the current time and closes the writer.
func cancelWrite(writer io.Writer) error {
	var errDeadline error
	var errCloser error

	if swd, ok := writer.(interface{ SetWriteDeadline(time.Time) error }); ok {
		errDeadline = swd.SetWriteDeadline(time.Now())
	}

	if closer, ok := writer.(io.Closer); ok {
		errCloser = closer.Close()
	}

	return errors.Join(errDeadline, errCloser)
}
