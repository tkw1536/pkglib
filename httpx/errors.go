//spellchecker:words httpx
package httpx

//spellchecker:words http
import (
	"fmt"
	"net/http"
)

//spellchecker:words nolint errname

// StatusCode represents an error based on a http status code.
// The integer is the http status code.
//
// StatusCode implements both [error] and [http.Handler].
// When used as a handler, it sets the appropriate status code, and returns a simple text response.
//
//nolint:errname
type StatusCode int

// check that StatusCode indeed implements error and http.Handler
var (
	_ error        = (StatusCode)(0)
	_ http.Handler = (StatusCode)(0)
)

// String returns the status text belonging to this error.
func (code StatusCode) String() string {
	return http.StatusText(int(code))
}

// GoString returns a go source code representation of string.
func (code StatusCode) GoString() string {
	return fmt.Sprintf("httpx.StatusCode(%d/* %s */)", code, code.String())
}

// Error implements the built-in [error] interface.
func (code StatusCode) Error() string {
	return "httpx: " + code.String()
}

// ServeHTTP implements [http.Handler].
func (code StatusCode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", ContentTypeText)
	w.WriteHeader(int(code))
	_, _ = w.Write([]byte(code.String()))
}

// Common Errors accepted by most httpx functions.
//
// These are guaranteed to implement both [error] and [http.Handler].
// See also [StatusCodeError].
const (
	ErrInternalServerError = StatusCode(http.StatusInternalServerError)
	ErrBadRequest          = StatusCode(http.StatusBadRequest)
	ErrNotFound            = StatusCode(http.StatusNotFound)
	ErrForbidden           = StatusCode(http.StatusForbidden)
	ErrMethodNotAllowed    = StatusCode(http.StatusMethodNotAllowed)
)
