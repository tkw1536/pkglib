// Package mux provides [Mux]
package mux

//spellchecker:words http
import (
	"net/http"
)

// Mux routes requests to different handlers.
// See the [Add] for how requests are matched to their handler.
type Mux struct {
	prefixes map[string][]handler
	exacts   map[string][]handler

	// NotFound is called when no prefix is matched.
	NotFound http.Handler
}

type handler struct {
	Predicate Predicate
	http.Handler
}

// Add adds a new handler to this Mux.
//
// path is the path or prefix to match for the given handler.
// exact determines if only the exact path is to be matched or the entire prefix.
// predicate is an additional predicate that must return true on any incoming request for the path to match.
func (mux *Mux) Add(path string, predicate Predicate, exact bool, h http.Handler) {
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
func (mux *Mux) Match(r *http.Request) (http.Handler, bool) {
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

// ServeHTTP serves requests to this mux.
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// find the right handler, or go into not found mode
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

// A Predicate is a function that matches a request.
// The nil predicate always matches.
type Predicate func(*http.Request) bool

// Call checks if this predicate matches the given request.
func (p Predicate) Call(r *http.Request) bool {
	if p == nil {
		return true
	}
	return p(r)
}
