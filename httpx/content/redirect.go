//spellchecker:words content
package content

//spellchecker:words http github pkglib httpx recovery
import (
	"net/http"

	"go.tkw01536.de/pkglib/httpx"
	"go.tkw01536.de/pkglib/recovery"
)

// Redirect creates a new [RedirectHandler] based on the given function.
// The Interceptor will be [httpx.TextInterceptor].
func Redirect(handler RedirectFunc) RedirectHandler {
	return RedirectHandler{
		Handler:     handler,
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
type RedirectFunc func(r *http.Request) (location string, code int, err error)

// RedirectHandler is a [http.Handler] that redirects every request based on the result of invoking Handler.
type RedirectHandler struct {
	Handler     RedirectFunc
	Interceptor httpx.ErrInterceptor
}

// ServeHTTP calls r(r) and returns json.
func (rh RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// call the function
	url, code, err := recovery.Safe2(func() (string, int, error) { return rh.Handler(r) })

	// intercept the errors
	if rh.Interceptor.Intercept(w, r, err) {
		return
	}

	// do the redirect
	http.Redirect(w, r, url, code)
}
