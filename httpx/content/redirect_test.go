//spellchecker:words content
package content_test

//spellchecker:words context http httptest github pkglib httpx content
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/httpx/content"
)

func ExampleRedirect() {

	// create a redirect based on the url
	handler := content.Redirect(func(r *http.Request) (string, int, error) {
		switch r.URL.Path {
		case "/temporary.example":
			return "https://example.com/", http.StatusTemporaryRedirect, nil
		case "/permanent.example":
			return "https://example.com/", http.StatusPermanentRedirect, nil
		case "/notfound":
			return "", 0, httpx.ErrNotFound
		}
		panic("never reached")
	})

	makeRequest := func(path string) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
		if err != nil {
			panic(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		rrr := rr.Result()
		body, _ := io.ReadAll(rrr.Body)
		fmt.Printf("%s returned code %d with location header %q and body %q\n", path, rrr.StatusCode, rrr.Header.Get("Location"), string(body))
	}

	// invoke the handler a bunch of times
	makeRequest("/temporary.example")
	makeRequest("/permanent.example")
	makeRequest("/notfound")
	// Output: /temporary.example returned code 307 with location header "https://example.com/" and body "<a href=\"https://example.com/\">Temporary Redirect</a>.\n\n"
	// /permanent.example returned code 308 with location header "https://example.com/" and body "<a href=\"https://example.com/\">Permanent Redirect</a>.\n\n"
	// /notfound returned code 404 with location header "" and body "Not Found"
}
