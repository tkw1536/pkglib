package mux_test

//spellchecker:words context http httptest github pkglib httpx
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/httpx/mux"
)

func ExampleMux() {
	var mux mux.Mux
	mux.NotFound = httpx.ErrNotFound

	// handle the "/something" route
	mux.Add("/something", nil, false, httpx.Response{Body: []byte("/something")})

	// handle the post route only for post calls
	mux.Add("/something/post", func(r *http.Request) bool { return r.Method == http.MethodPost }, false, httpx.Response{Body: []byte("/post")})

	// handle /something/exact route only on exact
	mux.Add("/something/exact", nil, true, httpx.Response{Body: []byte("/exact")})

	// a function to make a request to a specific method
	makeRequest := func(method, path string) {
		req, err := http.NewRequestWithContext(context.Background(), method, path, nil)
		if err != nil {
			panic(err)
		}

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		rrr := rr.Result()
		result, _ := io.ReadAll(rrr.Body)
		fmt.Printf("%s %q returned code %d with %s %q\n", method, path, rrr.StatusCode, rrr.Header.Get("Content-Type"), string(result))
	}

	makeRequest(http.MethodGet, "/something")
	makeRequest(http.MethodGet, "/something/post")
	makeRequest(http.MethodPost, "/something/post")
	makeRequest(http.MethodGet, "/something/exact")
	makeRequest(http.MethodGet, "/something/exact/sub")
	makeRequest(http.MethodGet, "/notfound")

	// Output: GET "/something" returned code 200 with text/plain; charset=utf-8 "/something"
	// GET "/something/post" returned code 200 with text/plain; charset=utf-8 "/something"
	// POST "/something/post" returned code 200 with text/plain; charset=utf-8 "/post"
	// GET "/something/exact" returned code 200 with text/plain; charset=utf-8 "/exact"
	// GET "/something/exact/sub" returned code 200 with text/plain; charset=utf-8 "/something"
	// GET "/notfound" returned code 404 with text/plain; charset=utf-8 "Not Found"
}
