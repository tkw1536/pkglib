//spellchecker:words contextx
package contextx

//spellchecker:words context time
import (
	"context"
	"time"
)

// Anyways behaves similar to [context.WithTimeout].
// However if the context is already cancelled before Anyways is called, the returned context's Done() channel is only closed after timeout.
func Anyways(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// context is not yet cancelled => return as-is
	if err := parent.Err(); err == nil {
		return context.WithTimeout(parent, timeout)
	}

	// context is cancelled => create a new one with a custom timeout
	return context.WithTimeout(context.WithoutCancel(parent), timeout)
}
