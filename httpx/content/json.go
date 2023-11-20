package content

import (
	"encoding/json"
	"net/http"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/recovery"
)

// spellchecker: words httpx

// JSON creates a new [JSONHandler] based on the given function.
// The Interceptor will be [httpx.JSONInterceptor].
func JSON[T any](f func(r *http.Request) (T, error)) JSONHandler[T] {
	return JSONHandler[T]{
		Handler:     f,
		Interceptor: httpx.JSONInterceptor,
	}
}

// WriteJSON writes a JSON response of type T to w.
// If an error occurred, [httpx.JSONInterceptor] is used instead.
func WriteJSON[T any](result T, err error, w http.ResponseWriter, r *http.Request) {
	writeJSON(result, err, httpx.JSONInterceptor, w, r)
}

// JSONHandler implements [http.Handler] by marshaling values as json to the caller.
// In case of an error, a generic "internal server error" message is returned.
type JSONHandler[T any] struct {
	Handler     func(r *http.Request) (T, error)
	Interceptor httpx.ErrInterceptor
}

// ServeHTTP calls j(r) and returns json
func (j JSONHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := recovery.Safe(func() (T, error) { return j.Handler(r) })
	writeJSON(result, err, j.Interceptor, w, r)
}

func writeJSON[T any](result T, err error, interceptor httpx.ErrInterceptor, w http.ResponseWriter, r *http.Request) {
	// handle any errors
	if httpx.JSONInterceptor.Intercept(w, r, err) {
		return
	}

	// write out the response as json
	w.Header().Set("Content-Type", httpx.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
