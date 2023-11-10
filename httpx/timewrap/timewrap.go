// package timewrap provides a means of wrapping a http.Handler with facilities to measure time since a request started.
package timewrap

// spellchecker:words timewrap

import (
	"context"
	"net/http"
	"time"

	"github.com/tkw1536/pkglib/httpx"
)

type tRequestStart struct{}

// key used for storing request start
var requestStartKey = tRequestStart{}

// Wrap wraps an [http.Handler] and records the time each request starts
func Wrap(h http.Handler) http.Handler {
	return httpx.ContextHandler{
		Handler: h,
		Replacer: func(r *http.Request) context.Context {
			return context.WithValue(r.Context(), requestStartKey, time.Now())
		},
	}
}

// Start returns the time that the request r was started.
// If no time is stored, returns the current time.
func Start(r *http.Request) time.Time {
	if r == nil {
		return time.Now()
	}

	start := r.Context().Value(requestStartKey)
	if start == nil {
		return time.Now()
	}

	t, ok := start.(time.Time)
	if !ok {
		return time.Now()
	}
	return t
}

// Start returns the time since the request r was started.
// If no time is stored in the request, returns 0.
func Since(r *http.Request) time.Duration {
	if r == nil {
		return 0
	}

	start := r.Context().Value(requestStartKey)
	if start == nil {
		return 0
	}

	t, ok := start.(time.Time)
	if !ok {
		return 0
	}

	return time.Since(t)
}
