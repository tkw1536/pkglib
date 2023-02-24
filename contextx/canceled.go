package contextx

import (
	"context"
)

// Canceled returns a new context that has been canceled.
func Canceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
