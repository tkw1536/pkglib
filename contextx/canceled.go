//spellchecker:words contextx
package contextx

//spellchecker:words context errors
import (
	"context"
	"errors"
	"sync"
)

var errCanceled = errors.New("contextx.Canceled")

// ErrCanceled is the cancel cause returned by Canceled.
var ErrCanceled = errors.Join(context.Canceled, errCanceled)

var canceled = sync.OnceValue(func() context.Context {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(ErrCanceled)
	return ctx
})

// Canceled returns a non-nil, empty Context.
// It has no deadline, has no values, and is already canceled.
// Calling [Cause] returns [ErrCanceled].
//
// Canceled may or may not returned the same context for different invocations.
func Canceled() context.Context {
	return canceled()
}
