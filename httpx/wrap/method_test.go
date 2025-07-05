//spellchecker:words wrap
package wrap_test

//spellchecker:words context http httptest github pkglib httpx wrap
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"go.tkw01536.de/pkglib/httpx/wrap"
)

func ExampleMethods() {
	handler := wrap.Methods(
		// Create a new handler that echoes the appropriate method
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(r.Method))
		}),

		// and permit only the GET and POST methods
		http.MethodGet, http.MethodPost,
	)

	// A simple function to make a request with a specific method
	makeRequest := func(method string) {
		req, err := http.NewRequestWithContext(context.Background(), method, "/", nil)
		if err != nil {
			panic(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		result, _ := io.ReadAll(rr.Result().Body)
		fmt.Printf("%s returned code %d with content %q\n", method, rr.Result().StatusCode, string(result))
	}

	// and do a couple of the requests
	makeRequest(http.MethodGet)
	makeRequest(http.MethodPost)
	makeRequest(http.MethodHead)

	// Output: GET returned code 200 with content "GET"
	// POST returned code 200 with content "POST"
	// HEAD returned code 405 with content "Method Not Allowed"
}
