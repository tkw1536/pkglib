//spellchecker:words httpx
package httpx_test

//spellchecker:words context embed http httptest github pkglib httpx recovery
import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/recovery"
)

// Render an error page in response to a panic.
func ExampleRenderErrorPage() {
	// response for errors
	res := httpx.Response{StatusCode: http.StatusNotFound, Body: []byte("not found")}

	// a handler with an error page
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// do some operation, here it always fails for testing.
		err := fmt.Errorf("for debugging: %w", httpx.ErrNotFound)
		if err != nil {
			// render the error page
			httpx.RenderErrorPage(err, res, w, r)
			return
		}

		// ... do some normal processing here ...
		panic("normal rendering, never reached")
	})

	// run the request
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		if err != nil {
			panic(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		result := rr.Result()

		fmt.Printf("Got status: %d\n", result.StatusCode)
		fmt.Printf("Got content-type: %s\n", result.Header.Get("Content-Type"))
	}

	// Output: Got status: 404
	// Got content-type: text/html; charset=utf-8
}

// Render an error page in response to a panic.
func ExampleRenderErrorPage_panic() {
	// response for errors
	res := httpx.Response{StatusCode: http.StatusInternalServerError, Body: []byte("something went wrong")}

	// a handler with an error page
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recover any errors
			if err := recovery.Recover(recover()); err != nil {
				// render the error page
				// which will replace it with a text/html page
				httpx.RenderErrorPage(err, res, w, r)
			}
		}()

		// ... do actual code ...
		panic("error for testing")
	})

	// run the request
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		if err != nil {
			panic(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		result := rr.Result()

		fmt.Printf("Got status: %d\n", result.StatusCode)
		fmt.Printf("Got content-type: %s\n", result.Header.Get("Content-Type"))
	}

	// Output: Got status: 500
	// Got content-type: text/html; charset=utf-8
}
