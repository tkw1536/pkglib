package httpx

import (
	"context"
	"net/http"
)

// WithContextWrapper creates a new ContextHandler, wrapping handler and replacing the context using f.
// If f is nil, or returns a nil context, the incoming request is forwarded as is.
func WithContextWrapper(handler http.Handler, f func(context.Context) context.Context) ContextHandler {
	// NOTE(twiesing): This function is untested, because ContextHandler is tested.
	replacer := func(r *http.Request) context.Context { return f(r.Context()) }
	if f == nil {
		replacer = nil
	}

	return ContextHandler{
		Handler:  handler,
		Replacer: replacer,
	}
}

// ContextHandler wraps Handler, replacing every context using the replacer function.
type ContextHandler struct {
	// Handler is the handler this ContextHandler wraps.
	Handler http.Handler

	// Replacer is called to replace the context of any incoming request.
	// If Replacer is nil, or returns a nil context, the context of the incoming request is left unchanged.
	Replacer func(*http.Request) context.Context
}

// ServeHTTP modifies the context of an incoming request, and passes it to Handler.
func (ch ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ch.Replacer != nil {
		ctx := ch.Replacer(r)
		if ctx != nil {
			r = r.WithContext(ctx)
		}
	}

	ch.Handler.ServeHTTP(w, r)
}
