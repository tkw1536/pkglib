package content_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/httpx/content"
)

func ExampleJSON() {

	// create a redirect based on the url
	handler := content.JSON(func(r *http.Request) (any, error) {
		switch r.URL.Path {
		case "/value":
			return 69, nil
		case "/slice":
			return []any{"hello", 42}, nil
		case "/notfound":
			return nil, httpx.ErrNotFound
		}
		panic("other error")
	})

	makeRequest := func(path string) {
		req, err := http.NewRequest(http.MethodGet, path, nil)
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
	makeRequest("/value")
	makeRequest("/slice")
	makeRequest("/notfound")
	makeRequest("/other")

	// Output: /value returned code 200 with location header "" and body "69\n"
	// /slice returned code 200 with location header "" and body "[\"hello\",42]\n"
	// /notfound returned code 404 with location header "" and body "{\"code\":404,\"status\":\"Not Found\"}"
	// /other returned code 500 with location header "" and body "{\"code\":500,\"status\":\"Internal Server Error\"}"
}
