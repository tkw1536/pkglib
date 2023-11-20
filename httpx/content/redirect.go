package content

import (
	"net/http"

	"github.com/tkw1536/pkglib/httpx"
)

// spellchecker:words httpx

// Redirect creates a new [RedirectHandler] based on the given function.
// The Interceptor will be [httpx.TextInterceptor].
func Redirect(Handler RedirectFunc) RedirectHandler {
	return RedirectHandler{
		Handler:     Handler,
		Interceptor: httpx.TextInterceptor,
	}
}

// RedirectFunc is invoked with an http.Request to determine the redirect behavior for a specific function.
//
// location should be the destination of the redirect.
// code should indicate the type of redirect, typically one of [http.StatusFound], [http.StatusTemporaryRedirect] or [http.StatusPermanentRedirect].
// error indicates any error that occurred.
//
// If error is non-nil it is intercepted by an appropriate [httpx.ErrInterceptor].
type RedirectFunc = func(r *http.Request) (location string, code int, err error)

// RedirectHandler is a [http.Handler] that redirects every request based on the result of invoking Handler.
type RedirectHandler struct {
	Handler     RedirectFunc
	Interceptor httpx.ErrInterceptor
}

// ServeHTTP calls r(r) and returns json
func (rh RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// call the function
	url, code, err := rh.Handler(r)

	// intercept the errors
	if rh.Interceptor.Intercept(w, r, err) {
		return
	}

	// do the redirect
	http.Redirect(w, r, url, code)
}
