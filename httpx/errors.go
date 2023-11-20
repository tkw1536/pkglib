package httpx

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// StatusCode represents an error based on a http status code.
// The integer is the http status code.
//
// StatusCode implements [error] as well as [http.Handler].
// When used as a handler, it sets the appropriate status code, and returns a simple text response.
type StatusCode int

// String returns the status text belonging to this error.
func (code StatusCode) String() string {
	return http.StatusText(int(code))
}

// GoString returns a go source code representation of string.
func (code StatusCode) GoString() string {
	return fmt.Sprintf("httpx.StandardError(%d/* %s */)", code, code.String())
}

// Error implements the built-in [error] interface.
func (code StatusCode) Error() string {
	return "httpx: " + code.String()
}

// ServeHTTP implements [http.Handler].
func (code StatusCode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", ContentTypeText)
	w.WriteHeader(int(code))
	w.Write([]byte(code.String()))
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

// check that http.Handler indeed implements error and http.Handler
var (
	_ error        = (StatusCode)(0)
	_ http.Handler = (StatusCode)(0)
)

// Recover returns an error that represents an error caught from recover.
// When passed nil, returns nil.
//
// It should be used as:
//
//	if err := Recover(recover()); err != nil {
//		// ... handle here ...
//	}
func Recover(value any) error {
	if value == nil {
		return nil
	}
	return recovery{
		Stack: debug.Stack(),
		Value: value,
	}
}

type recovery struct {
	Stack []byte
	Value any
}

func (r recovery) GoString() string {
	return fmt.Sprintf("httpx.recovery{/* recover() = %#v */}", r.Value)
}

func (r recovery) Error() string {
	return fmt.Sprintf("%v\n\n%s", r.Value, r.Stack)
}
