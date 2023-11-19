package httpx_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tkw1536/pkglib/httpx"
)

// ErrInterceptor
func ExampleErrInterceptor() {
	// create an error interceptor
	interceptor := httpx.ErrInterceptor{

		// handle [ErrNotFound] with a not found response
		Errors: map[error]httpx.Response{
			// error not found (and wraps thereof) return that status code
			httpx.ErrNotFound: {
				StatusCode: http.StatusNotFound,
				Body:       []byte("Not Found"),
			},

			// forbidden (isn't actually used in this example)
			httpx.ErrForbidden: {
				StatusCode: http.StatusForbidden,
				Body:       []byte("Forbidden"),
			},
		},

		// fallback to a generic not found error
		Fallback: httpx.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       []byte("fallback error"),
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ... do some work ...
		// in prod this would be an error returned from some operation
		result := map[string]error{
			"/":         nil, // no error
			"/notfound": httpx.ErrNotFound,
			"/wrapped":  fmt.Errorf("wrapped: %w", httpx.ErrNotFound),
		}[r.URL.Path]

		// intercept an error
		if interceptor.Intercept(w, r, result) {
			return
		}

		w.Write([]byte("Normal response"))
	})

	// a function to make a request to a specific method
	makeRequest := func(path string) {
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			panic(err)
		}

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		rrr := rr.Result()
		result, _ := io.ReadAll(rrr.Body)
		fmt.Printf("%q returned code %d with %s %q\n", path, rrr.StatusCode, rrr.Header.Get("Content-Type"), string(result))
	}

	makeRequest("/")
	makeRequest("/notfound")
	makeRequest("/wrapped")

	// Output: "/" returned code 200 with text/plain; charset=utf-8 "Normal response"
	// "/notfound" returned code 404 with text/plain; charset=utf-8 "Not Found"
	// "/wrapped" returned code 404 with text/plain; charset=utf-8 "Not Found"
}
