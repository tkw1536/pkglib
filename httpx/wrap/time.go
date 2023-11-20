package wrap

// spellchecker:words timewrap

import (
	"context"
	"net/http"
	"time"
)

// requestTime is a type for storing the request time
type requestTime struct{}

var requestTimeKey = requestTime{}

// Time wraps an [http.Handler], storing the time a request was started within it.
// To retrieve stored time, see [TimeStart] and [TimeSince].
func Time(h http.Handler) http.Handler {
	return Context(h, func(r *http.Request) (context.Context, context.CancelFunc) {
		return context.WithValue(r.Context(), requestTimeKey, time.Now()), nil
	})
}

// TimeStart returns the time that the request r was started.
// Must be called from within a handler wrapped with [Time].
// If no time is stored, returns the current time.
func TimeStart(r *http.Request) time.Time {
	if r == nil {
		return time.Now()
	}

	start := r.Context().Value(requestTimeKey)
	if start == nil {
		return time.Now()
	}

	t, ok := start.(time.Time)
	if !ok {
		return time.Now()
	}
	return t
}

// TimeSince returns the time since the request r was started.
// Must be called from within a handler wrapped with [Time] to be effective.
// If no time is stored in the request, returns 0.
func TimeSince(r *http.Request) time.Duration {
	if r == nil {
		return 0
	}

	start := r.Context().Value(requestTimeKey)
	if start == nil {
		return 0
	}

	t, ok := start.(time.Time)
	if !ok {
		return 0
	}

	return time.Since(t)
}
