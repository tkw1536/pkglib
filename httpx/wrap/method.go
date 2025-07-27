//spellchecker:words wrap
package wrap

//spellchecker:words http pkglib httpx
import (
	"net/http"

	"go.tkw01536.de/pkglib/httpx"
)

// Methods wraps handler, rejecting requests not using any of the provided methods.
// Requests with rejected methods return a generic "Method Not Allowed" response with appropriate status code.
func Methods(handler http.Handler, methods ...string) http.Handler {
	// create a map of allowed methods
	allowed := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		allowed[method] = struct{}{}
	}

	// create an appropriate handler function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; !ok {
			methodNotAllowed.ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// methodNotAllowed is the response returned when there is no response.
var methodNotAllowed = httpx.Response{
	StatusCode: http.StatusMethodNotAllowed,
	Body:       []byte("Method Not Allowed"),
}
