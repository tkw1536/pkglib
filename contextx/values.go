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
