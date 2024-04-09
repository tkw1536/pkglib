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

			_, _ = w.Write([]byte(content))
		}),

		// Wrap it using a function that automatically sets the key
		func(r *http.Request) (context.Context, context.CancelFunc) {
			return context.WithValue(r.Context(), responseKey, "this response got set in a wrapper"), nil
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

func ExampleContext_cancel() {
	handler := wrap.Context(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("calling handler")
			_, _ = w.Write(nil)
		}),

		// Wrap it using a function that automatically sets the key
		func(r *http.Request) (context.Context, context.CancelFunc) {
			return r.Context(), func() { fmt.Printf("calling CancelFunc") }
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

	// Output: calling handler
	// calling CancelFunc
}
