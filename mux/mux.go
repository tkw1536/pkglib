// Package mux provides mux
package mux

import (
	"context"
	"net/http"
	"runtime/debug"
)

// TODO: TESTME

// Mux handles different requests using either exact or prefix path matching.
// Each request is provided with a context.
type Mux[C any] struct {
	prefixes map[string][]handler
	exacts   map[string][]handler

	// Context adds context to the provided request.
	// If non-nil it is called exactly once for each request.
	// When context is nil, the zero value of type C is added as a context.
	Context func(r *http.Request) C

	// Panic, if non-nil, is called when a panic occurs in any step of the response process.
	// Additionally the stack trace right after recover is provided.
	// When Panic is nil, no recover is performed.
	Panic func(panic any, stack []byte, w http.ResponseWriter, r *http.Request)

	// NotFound is called when a specific path cannot be associated to a handler.
	NotFound http.Handler
}

// muxContextKey returns the context type of a mux
type muxContextKey struct{}

var muxKey = muxContextKey{}

type handler struct {
	Predicate Predicate
	http.Handler
}

// Prepare prepares a request object to be passed to a user-defined handler.
// This calls the context function (if non-nil) and adds it to the context.
func (mux *Mux[T]) Prepare(r *http.Request) *http.Request {
	if mux == nil || mux.Context == nil {
		return r
	}

	ctx := context.WithValue(r.Context(), muxKey, mux.Context(r))
	return r.WithContext(ctx)
}

// ContextOf returns the context object belonging to the provided context.
// If no context object exists, returns the zero value.
func (mux *Mux[T]) ContextOf(r *http.Request) (t T) {
	value, ok := r.Context().Value(muxKey).(T)
	if !ok {
		return t
	}
	return value
}

// HandlerOptions are options for a single handler.
// The zero-value is ready to use.
type HandlerOptions struct {
	Exact bool

	// Pre
	Predicate Predicate

	// Priority handles priorities with the same matched prefix.
	// Within the same prefix, handlers with the same priority are matched first.
	// If two handlers have the same priority, the one that was added first will be called.
	Priority int
}

// Add adds a new handler to this Mux.
//
// path is the path or prefix to match for the given handler.
// exact determines if only the exact path is to be matched or the entire prefix.
// predicate is an additional predicate that must return true on any incoming request for the path to match.
func (mux *Mux[T]) Add(path string, predicate Predicate, exact bool, h http.Handler) {
	if mux.exacts == nil {
		mux.exacts = make(map[string][]handler)
	}
	if mux.prefixes == nil {
		mux.prefixes = make(map[string][]handler)
	}

	mPath := NormalizePath(path)
	mHandler := handler{Predicate: predicate, Handler: h}
	if exact {
		mux.exacts[mPath] = append(mux.exacts[mPath], mHandler)
	} else {
		mux.prefixes[mPath] = append(mux.prefixes[mPath], mHandler)
	}
}

// Match returns the handler to be applied for the given request.
// Match expects that Prepare() has been called on the given request.
func (mux *Mux[T]) Match(r *http.Request) (http.Handler, bool) {
	if mux == nil {
		return nil, false
	}

	candidate := NormalizePath(r.URL.Path)

	// match the exact path first
	for _, h := range mux.exacts[candidate] {
		if h.Predicate.Call(r) {
			return h.Handler, true
		}
	}

	// iterate over path segment candidates
	for {
		// check the current candidate
		for _, h := range mux.prefixes[candidate] {
			if h.Predicate.Call(r) {
				return h.Handler, true
			}
		}

		// if the candidate is the root url, we can bail out now
		if len(candidate) == 0 || candidate == "/" {
			return nil, false
		}

		// move to the parent segment
		candidate = parentSegment(candidate)
	}

}

// ServeHTTP serves requests to this mux
func (mux *Mux[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// handle panics with the panic handler
	defer func() {
		// if there is no panic handler, don't do anything
		if mux == nil || mux.Panic == nil {
			return
		}

		// try to recover
		caught := recover()
		if caught == nil {
			return
		}

		// call the panic handler
		mux.Panic(caught, debug.Stack(), w, r)
	}()

	// prepare the request
	r = mux.Prepare(r)

	// find the right handler
	// or go into 404 mode
	handler, ok := mux.Match(r)
	if !ok {
		if mux == nil || mux.NotFound == nil {
			http.NotFound(w, r)
			return
		}
		mux.NotFound.ServeHTTP(w, r)
		return
	}

	// call the actual handling
	handler.ServeHTTP(w, r)
}

// Predicate represents a matching predicate for a given request.
// The nil predicate always matches.
type Predicate func(r *http.Request) bool

// Call checks if this predicate matches the given request.
func (p Predicate) Call(r *http.Request) bool {
	if p == nil {
		return true
	}
	return p(r)
}
