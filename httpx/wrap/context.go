// Package wrap provides wrappers for [http.Handler]s.
package wrap

import (
	"context"
	"net/http"
)

// ContextFunc is a function that replaces contexts for a given request.
// A nil ContextFunc leaves the original context intact.
//
// - the returned context, if non-nil, is used to replace the context of the request.
// - the returned CancelFunc is called once the request ends.
type ContextFunc = func(r *http.Request) (context.Context, context.CancelFunc)

// Context wraps handler, replacing the context of any request using the given function.
func Context(handler http.Handler, f ContextFunc) http.Handler {
	if f == nil {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// compute the context
		ctx, cancel := f(r)

		// call cancel once the request ends
		if cancel != nil {
			defer cancel()
		}

		// use the new context
		if ctx != nil {
			r = r.WithContext(ctx)
		}

		// do the handling
		handler.ServeHTTP(w, r)
	})
}
