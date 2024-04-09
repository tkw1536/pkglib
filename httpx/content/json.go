package content

import (
	"encoding/json"
	"net/http"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/recovery"
)

// spellchecker: words httpx jsoni

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
func WriteJSON[T any](result T, err error, w http.ResponseWriter, r *http.Request) error {
	return WriteJSONI(result, err, httpx.JSONInterceptor, w, r)
}

// WriteJSONI is like [WriteJSON] but uses a custom interceptor.
func WriteJSONI[T any](result T, err error, interceptor httpx.ErrInterceptor, w http.ResponseWriter, r *http.Request) error {
	// handle any errors
	if interceptor.Intercept(w, r, err) {
		return nil
	}

	// write out the response as json
	w.Header().Set("Content-Type", httpx.ContentTypeJSON)
	{
		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return err
	}
}

// JSONHandler implements [http.Handler] by marshaling values as json to the caller.
// In case of an error, a generic "internal server error" message is returned.
type JSONHandler[T any] struct {
	Handler func(r *http.Request) (T, error)

	Interceptor        httpx.ErrInterceptor
	LogJSONEncodeError func(r *http.Request, err error)
}

// ServeHTTP calls j(r) and returns json
func (j JSONHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := recovery.Safe(func() (T, error) { return j.Handler(r) })
	{
		err := WriteJSONI(result, err, j.Interceptor, w, r)
		if err != nil && j.LogJSONEncodeError != nil {
			j.LogJSONEncodeError(r, err)
		}
	}
}
