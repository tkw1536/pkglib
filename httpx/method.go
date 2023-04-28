package httpx

import "net/http"

// PermitMethods returns a new http.Handler that calls handler for the given methods,
// and returns a generic "Method Not Allowed" response otherwise.
func PermitMethods(handler http.Handler, methods ...string) http.Handler {
	var wrapper permitMethods

	wrapper.Handler = handler
	wrapper.Methods = make(map[string]struct{}, len(methods))
	for _, method := range methods {
		wrapper.Methods[method] = struct{}{}
	}

	return wrapper
}

var methodNotAllowed = Response{
	StatusCode: http.StatusMethodNotAllowed,
	Body:       []byte("Method not allowed"),
}

type permitMethods struct {
	Handler http.Handler
	Methods map[string]struct{}
}

func (p permitMethods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := p.Methods[r.Method]; !ok {
		methodNotAllowed.ServeHTTP(w, r)
		return
	}
	p.Handler.ServeHTTP(w, r)
}
