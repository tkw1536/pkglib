// Package httpx provides additional [http.Handler]s and utility functions
//
//spellchecker:words httpx
package httpx_test

//spellchecker:words http httptest time github pkglib httpx
import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/tkw1536/pkglib/httpx"
)

//spellchecker:words modtime

// Using a response with a plain http status.
func ExampleResponse() {
	response := httpx.Response{
		StatusCode:  http.StatusOK,
		ContentType: httpx.ContentTypeHTML,
		Body:        []byte("<!DOCTYPE html>Hello world"),
	}

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}
	rr := httptest.NewRecorder()
	response.ServeHTTP(rr, req)

	result := rr.Result()
	body, _ := io.ReadAll(result.Body)

	fmt.Printf("Got status: %d\n", result.StatusCode)
	fmt.Printf("Got content-type: %s\n", result.Header.Get("Content-Type"))
	fmt.Printf("Got body: %s", string(body))

	// Output: Got status: 200
	// Got content-type: text/html; charset=utf-8
	// Got body: <!DOCTYPE html>Hello world
}

// It is possible to omit everything, and defaults will be set correctly.
func ExampleResponse_defaults() {
	response := httpx.Response{
		Body: []byte("Hello world"),
	}

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}
	rr := httptest.NewRecorder()
	response.ServeHTTP(rr, req)

	result := rr.Result()
	body, _ := io.ReadAll(result.Body)

	fmt.Printf("Got status: %d\n", result.StatusCode)
	fmt.Printf("Got content-type: %s\n", result.Header.Get("Content-Type"))
	fmt.Printf("Got body: %s", string(body))

	// Output: Got status: 200
	// Got content-type: text/plain; charset=utf-8
	// Got body: Hello world
}

// This means that appropriate 'if-modified-since' headers are respected.
func ExampleResponse_Now() {
	response := httpx.Response{
		Body: []byte("Hello world"),
	}.Now()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("If-Modified-Since", response.Modtime.Add(time.Second).Format(http.TimeFormat)) // set an if-modified-since

	rr := httptest.NewRecorder()
	response.ServeHTTP(rr, req)

	result := rr.Result()
	fmt.Printf("Got status: %d\n", result.StatusCode)

	// Output: Got status: 304
}
