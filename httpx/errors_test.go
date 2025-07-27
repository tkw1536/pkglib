//spellchecker:words httpx
package httpx_test

//spellchecker:words context http httptest pkglib httpx
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"go.tkw01536.de/pkglib/httpx"
)

func ExampleStatusCode() {
	// we can create an error based on the status code
	handler := httpx.StatusCode(http.StatusNotFound)

	// which automatically generates error messages
	fmt.Printf("String: %s\n", handler.String())
	fmt.Printf("GoString: %s\n", handler.GoString())
	fmt.Printf("Error: %s\n", handler.Error())

	// it also implements a static http.Handler
	{
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		if err != nil {
			panic(err)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		rrr := rr.Result()
		body, _ := io.ReadAll(rrr.Body)

		fmt.Printf("ServeHTTP() status: %d\n", rrr.StatusCode)
		fmt.Printf("ServeHTTP() content-type: %s\n", rrr.Header.Get("Content-Type"))
		fmt.Printf("ServeHTTP() body: %s\n", body)
	}

	// Output: String: Not Found
	// GoString: httpx.StatusCode(404/* Not Found */)
	// Error: httpx: Not Found
	// ServeHTTP() status: 404
	// ServeHTTP() content-type: text/plain; charset=utf-8
	// ServeHTTP() body: Not Found
}
