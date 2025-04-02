//spellchecker:words contextx
package contextx

//spellchecker:words context
import (
	"context"
)

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

// WithValuesOf creates a new context that inherits from parent, but values stored in values take precedence over already associated values.
// If a value is not found in values, the parent context is searched.
// For explicitly associating a specific map of values see [WithValues].
func WithValuesOf(parent, values context.Context) context.Context {
	return &valuesOf{
		Context: parent,
		values:  values,
	}
}

//nolint:containedctx
type valuesOf struct {
	context.Context
	values context.Context
}

func (vv *valuesOf) Value(key any) any {
	if value := vv.values.Value(key); value != nil {
		return value
	}
	return vv.Context.Value(key)
}
