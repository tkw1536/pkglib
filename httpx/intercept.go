//spellchecker:words httpx
package httpx

//spellchecker:words encoding json errors http
import (
	"encoding/json"
	"errors"
	"net/http"
)

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

// statuses intercepted by StatusInterceptor
var statuses = []StatusCode{
	ErrInternalServerError,
	ErrBadRequest,
	ErrNotFound,
	ErrForbidden,
	ErrMethodNotAllowed,
}

// commonInterceptor creates a new ErrInterceptor handling default responses.
// If body returns err != nil, commonInterceptor calls panic().
func commonInterceptor(contentType string, body func(code StatusCode) []byte) ErrInterceptor {
	var interceptor ErrInterceptor

	interceptor.Errors = make(map[error]Response, len(statuses))
	for _, code := range statuses {
		interceptor.Errors[code] = Response{
			ContentType: contentType,
			StatusCode:  int(code),
			Body:        body(code),
		}.Minify()
	}

	interceptor.Fallback = interceptor.Errors[ErrInternalServerError]

	return interceptor
}

// Common interceptors for specific content types.
//
// These handle all common http status codes by sending their response with a common error code.
// See the Error constants of this package for supported errors.
var (
	TextInterceptor ErrInterceptor
	JSONInterceptor ErrInterceptor
	HTMLInterceptor ErrInterceptor
)

func init() {
	TextInterceptor = commonInterceptor(ContentTypeText, func(code StatusCode) []byte { return []byte(code.String()) })
	JSONInterceptor = commonInterceptor(ContentTypeJSON, func(code StatusCode) []byte {
		res, err := json.Marshal(map[string]any{"status": code.String(), "code": int(code)})
		if err != nil {
			panic(err)
		}
		return res
	})
	HTMLInterceptor = commonInterceptor(ContentTypeHTML, func(code StatusCode) []byte {
		cstring := code.String()
		return []byte(`<!DOCTYPE HTML><title>` + cstring + `</title>` + cstring)
	})
}
