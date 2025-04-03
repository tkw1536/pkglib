//spellchecker:words content
package content_test

//spellchecker:words context html template http httptest testing github pkglib httpx content
import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/httpx/content"
)

type ValueContainer struct {
	Value any
}

func ExampleHTML() {

	var handler content.HTMLHandler[any]
	handler.Interceptor = httpx.HTMLInterceptor
	handler.Template = template.Must(template.New("example").Parse(`<!DOCTYPE html>Result: {{ .Value }}`))
	handler.Handler = func(r *http.Request) (any, error) {
		switch r.URL.Path {
		case "/value":
			return ValueContainer{69}, nil
		case "/slice":
			return ValueContainer{[]any{"hello", 42}}, nil
		case "/notfound":
			return ValueContainer{nil}, httpx.ErrNotFound
		case "/template_error":
			return 42, nil

		}
		panic("other error")
	}

	// invoke the handler a bunch of times
	fmt.Println(makeRequest(handler, "/value"))
	fmt.Println(makeRequest(handler, "/slice"))
	fmt.Println(makeRequest(handler, "/notfound"))
	fmt.Println(makeRequest(handler, "/other"))
	fmt.Println(makeRequest(handler, "/template_error"))

	// Output: /value returned code 200 with location header "" and body "<!doctype html>Result: 69"
	// /slice returned code 200 with location header "" and body "<!doctype html>Result: [hello 42]"
	// /notfound returned code 404 with location header "" and body "<!doctype html><title>Not Found</title>Not Found"
	// /other returned code 500 with location header "" and body "<!doctype html><title>Internal Server Error</title>Internal Server Error"
	// /template_error returned code 500 with location header "" and body "<!doctype html>Result:"
}

func makeRequest(handler http.Handler, path string) string {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		panic(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	rrr := rr.Result()
	body, _ := io.ReadAll(rrr.Body)
	return fmt.Sprintf("%s returned code %d with location header %q and body %q", path, rrr.StatusCode, rrr.Header.Get("Location"), string(body))
}

func TestHTML_LogJSONEncodeError(t *testing.T) {
	t.Parallel()

	var handler content.HTMLHandler[any]
	handler.Interceptor = httpx.HTMLInterceptor
	handler.Template = template.Must(template.New("example").Parse(`<!DOCTYPE html>Result: {{ .Value }}`))
	handler.Handler = func(r *http.Request) (any, error) {
		switch r.URL.Path {
		case "/ok":
			return ValueContainer{69}, nil
		case "/broken":
			return 42, nil

		}
		panic("other error")
	}

	for _, tt := range []struct {
		Path       string
		WantCalled bool
	}{
		{Path: "/ok", WantCalled: false},
		{Path: "/broken", WantCalled: true},
	} {
		t.Run(tt.Path, func(t *testing.T) {
			t.Parallel()

			// setup a LogTemplateError that records if it was called or not
			called := false
			handler.LogTemplateExecuteError = func(r *http.Request, err error) {
				called = true
			}

			makeRequest(handler, tt.Path)

			if called != tt.WantCalled {
				t.Errorf("want called = %t, got called = %t", tt.WantCalled, called)
			}
		})
	}
}
