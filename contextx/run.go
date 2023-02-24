package contextx

// TESTME

import (
	"context"
	"sync"
)

// Run adds context-like functionality to a function that only supports explicit cancellation functionality.
//
// In principle, it calls f and returns the provided returns f(), nil.
// If the context is cancelled before f returns, Run instead invokes cancel and returns f(), ctx.Err().
//
// f can control at which point cancellation may occur.
// It must call start as soon as cancel may be called.
//
// Calling start multiple times or not at all is also permitted.
// However these use-cases should be carefully considered.
// In cases where start is not called, cancel will never be called, regardless of when ctx is cancelled.
// In cases where start is called multiple times, cancel may be invoked immediatly after the first invocation.
//
// Run always waits for f to return, and always returns the return value of f as the first argument, even if cancel is called.
func Run[T any](ctx context.Context, f func(start func()) T, cancel func()) (t T, err error) {
	t, _, err = Run2(ctx, func(start func()) (T, struct{}) {
		return f(start), struct{}{}
	}, cancel)
	return
}

// Run2 behaves exactly like Run, except that it allows f to return two values.
func Run2[T1, T2 any](ctx context.Context, f func(start func()) (T1, T2), cancel func()) (t1 T1, t2 T2, err error) {
	// special case: context is already closed
	if err := ctx.Err(); err != nil {
		return t1, t2, err
	}

	cancelled := make(chan struct{}, 1)

	fdone := make(chan struct{})
	go func() {
		defer close(fdone)
		defer close(cancelled)

		var cancelOnce sync.Once

		// store the return values.
		t1, t2 = f(func() {
			cancelOnce.Do(func() {
				cancelled <- struct{}{}
			})
		})
	}()

	select {
	case <-fdone: // normal exit
	case <-ctx.Done(): // context was cancelled

		// call the cancel function once start() has been called
		_, ok := <-cancelled
		if ok {
			cancel()
		}

		// wait for the function to be done
		// and set the error
		<-fdone
		err = ctx.Err()
	}
	return
}
