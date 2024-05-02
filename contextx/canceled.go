//spellchecker:words contextx
package contextx

//spellchecker:words context errors
import (
	"context"
	"errors"
)

// ErrCanceled is the cancel cause returned by Canceled.
var ErrCanceled = errors.Join(context.Canceled, errors.New("contextx.Canceled"))

// Canceled returns a non-nil, empty Context.
// It has no deadline, has no values, and is already canceled.
// The cancel cause is ErrCanceled.
//
// Canceled may or may not return the same context for different invocations.
func Canceled() context.Context {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(ErrCanceled)
	return ctx
}
