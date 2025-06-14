//spellchecker:words contextx
package contextx

//spellchecker:words context
import (
	"context"
)

//spellchecker:words nolint containedctx

// WithValues creates a new context that inherits from parent, but has associated values from values.
//
// This function is equivalent to repeated invocations of [context.WithValue].
// See the appropriate documentation for details on restrictions of keys and values to be used.
func WithValues(parent context.Context, values map[any]any) context.Context {
	ctx := parent
	for key, val := range values {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

// WithValuesOf creates a new context that holds values store in values, but is canceled when parent is canceled.
// Any values stored only in parent are ignored.
func WithValuesOf(parent, values context.Context) context.Context {
	ctx, cancel := context.WithCancelCause(context.WithoutCancel(values))

	// forward the cancel cause of the child
	go func() {
		select {
		case <-parent.Done():
			cancel(context.Cause(ctx))
		case <-ctx.Done():
		}
	}()

	return ctx
}
