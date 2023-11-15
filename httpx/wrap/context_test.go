package wrap_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tkw1536/pkglib/httpx/wrap"
)

type responseKeyType struct{}

var responseKey = responseKeyType{}

func ExampleContext() {
	handler := wrap.Context(
		// Create a new handler that extracts a given context key
		// and writes it to the response
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			content, ok := r.Context().Value(responseKey).(string)
			if !ok {
				return
			}

			w.Write([]byte(content))
		}),

		// Wrap it using a function that automatically sets the key
		func(ctx context.Context) context.Context {
			return context.WithValue(ctx, responseKey, "this response got set in a wrapper")
		},
	)

	// create a new request
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}

	// serve the http request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Result().StatusCode; status != http.StatusOK {
		fmt.Println("Expected http.StatusOK")
	}

	result, _ := io.ReadAll(rr.Result().Body)
	fmt.Println(string(result))

	// Output: this response got set in a wrapper
}
