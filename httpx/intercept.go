package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/tkw1536/pkglib/minify"
)

// TODO: Testme

// ErrInterceptor can handle errors for http responses and render appropriate error responses.
type ErrInterceptor struct {
	// Errors determines which error classes are intercepted.
	//
	// Errors are compared using [errors.Is] is undetermined order.
	// This means that if an error that [errors.Is] for multiple keys,
	// the returned response may any of the values.
	Errors map[error]Response

	// Fallback is the response for errors that are not of any of the above error classes.
	Fallback Response

	// RenderError indicates that instead of intercepting an error regularly
	// a human-readable error page with the appropriate error code should be displayed.
	// See [ErrorPage] for documentation on the error page handler.
	//
	// This option should only be used during development, as it exposes potentially security-critical data.
	RenderError bool

	// OnFallback is called when an unknown error is intercepted.
	OnFallback func(*http.Request, error)
}

// Intercept intercepts the given error, and writes the response to the struct.
// A response is written to w if and only error is not nil.
// The return value indicates if error was nil and a response was written.
//
// A typical use of an Intercept should be as follows:
//
//	// get interceptor from somewhere
//	var ei ErrInterceptor
//	// perform an operation, intercept the error or bail out
//	result, err := SomeOperation()
//	if ei.Intercept(w, r, err) {
//		return
//	}
//
// // ... write result to the response ...
//
// The precise behavior of Intercept is documented inside [ErrInterceptor] itself.
func (ei ErrInterceptor) Intercept(w http.ResponseWriter, r *http.Request, err error) (intercepted bool) {
	if err == nil {
		return false
	}

	if ei.RenderError {
		ei.renderError(w, r, err)
		return true
	}

	ei.interceptError(w, r, err)
	return true
}

func (ei ErrInterceptor) renderError(w http.ResponseWriter, r *http.Request, err error) {
	res, ok := ei.match(err)
	if !ok && ei.OnFallback != nil {
		ei.OnFallback(r, err)
	}

	RenderErrorPage(err, res, w, r)
}

func (ei ErrInterceptor) interceptError(w http.ResponseWriter, r *http.Request, err error) {
	res, ok := ei.match(err)
	if !ok && ei.OnFallback != nil {
		ei.OnFallback(r, err)
	}
	res.ServeHTTP(w, r)
}

func (ei ErrInterceptor) match(err error) (Response, bool) {
	for target, res := range ei.Errors {
		if errors.Is(err, target) {
			return res, true
		}
	}
	return ei.Fallback, false
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

// Recover returns an error that represents an error caught from recover.
// It should be used as:
//
//	if err := Recover(recover()); err != nil {
//		// ... handle here ...
//	}
func Recover(value any) error {
	if value == nil {
		return nil
	}
	return errRecover{
		Stack: debug.Stack(),
		Value: value,
	}
}

type errRecover struct {
	Stack []byte
	Value any
}

func (er errRecover) GoString() string {
	return "httpx.errRecover{/*details omitted*/}"
}

func (er errRecover) Error() string {
	return fmt.Sprintf("%v\n\n%s", er.Value, er.Stack)
}

// Common errors accepted by all httpx handlers
var (
	ErrInternalServerError = errors.New("httpx: Internal Server Error")
	ErrBadRequest          = errors.New("httpx: Bad Request")
	ErrNotFound            = errors.New("httpx: Not Found")
	ErrForbidden           = errors.New("httpx: Forbidden")
	ErrMethodNotAllowed    = errors.New("httpx: Method Not Allowed")
)

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
