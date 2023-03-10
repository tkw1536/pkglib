package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tkw1536/pkglib/minify"
)

// ErrInterceptor intercepts errors and directly returns specific responses for them
type ErrInterceptor struct {
	Errors   map[error]Response
	Fallback Response

	// onFallback (if non-nil) is called when a fallback error is intercepted
	OnFallback func(*http.Request, error)
}

// Intercept attempts to intercept the given error.
// When err is nil, does nothing.
//
// When err is not nil, first attempts to find a static response in errors and respond with that.
// Otherwise it returns the Fallback response.
// intercepted indicates if some response was sent.
func (ei ErrInterceptor) Intercept(w http.ResponseWriter, r *http.Request, err error) (intercepted bool) {
	if err == nil {
		return false
	}

	res, ok := ei.Errors[err]
	if !ok {
		ei.onFallbackSafe(r, err)
		res = ei.Fallback
	}

	res.ServeHTTP(w, r)
	return true
}
func (ei ErrInterceptor) onFallbackSafe(req *http.Request, err error) {
	fallback := ei.OnFallback
	if fallback == nil {
		fallback = InterceptorOnFallback
	}

	if fallback == nil {
		return
	}

	defer func() {
		recover()
	}()

	fallback(req, err)
}

// StatusInterceptor creates a new ErrInterceptor handling default responses.
// If body returns err != nil, StatusInterceptor calls panic().
func StatusInterceptor(contentType string, body func(code int, text string) ([]byte, error)) ErrInterceptor {
	makeResponse := func(code int) (res Response) {
		var err error
		res.Body, err = body(code, http.StatusText(code))
		if err != nil {
			panic("StatusInterceptor: err != nil")
		}

		res.ContentType = contentType
		res.StatusCode = code
		return
	}

	return ErrInterceptor{
		Errors: map[error]Response{
			ErrInternalServerError: makeResponse(http.StatusInternalServerError),
			ErrBadRequest:          makeResponse(http.StatusBadRequest),
			ErrNotFound:            makeResponse(http.StatusNotFound),
			ErrForbidden:           makeResponse(http.StatusForbidden),
			ErrMethodNotAllowed:    makeResponse(http.StatusMethodNotAllowed),
		},
		Fallback: makeResponse(http.StatusInternalServerError),
	}
}

// Common errors accepted by all httpx handlers
var (
	ErrInternalServerError = errors.New("httpx: Internal Server Error")
	ErrBadRequest          = errors.New("httpx: Bad Request")
	ErrNotFound            = errors.New("httpx: Not Found")
	ErrForbidden           = errors.New("httpx: Forbidden")
	ErrMethodNotAllowed    = errors.New("httpx: Method Not Allowed")
)

// InterceptorOnFallback (if non-nil) is called used by any ErrInterceptor as default OnFallback.
// It is triggered whenever OnFallback is nil.
//
// This should be set by any client package to globally log any errors.
var InterceptorOnFallback func(*http.Request, error)

var (
	TextInterceptor = StatusInterceptor("text/plain", func(code int, text string) ([]byte, error) {
		return []byte(text), nil
	})
	JSONInterceptor = StatusInterceptor("application/json", func(code int, text string) ([]byte, error) {
		return json.Marshal(map[string]any{"status": text, "code": code})
	})
	HTMLInterceptor = StatusInterceptor("text/html", func(code int, text string) ([]byte, error) {
		return minify.MinifyBytes("text/html", []byte(`<!DOCTYPE HTML><title>`+text+`</title>`+text)), nil
	})
)
