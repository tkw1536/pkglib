// Package wrap provides wrappers for [http.Handler]s.
package wrap

import (
	"context"
	"net/http"
)

// Context wraps handler, replacing the context of any request using the given function
// Requests with rejected methods return a generic "Method Not Allowed" response.
func Context(handler http.Handler, f func(context.Context) context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(f(r.Context())))
	})
}
