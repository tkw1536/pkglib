//spellchecker:words content
package content_test

//spellchecker:words errors http testing pkglib httpx content
import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"go.tkw01536.de/pkglib/httpx"
	"go.tkw01536.de/pkglib/httpx/content"
)

type BrokenMarshalJSON struct{}

var errBrokenMarshalJSON = errors.New("BrokenMarshalJSON.MarshalJSON error")

func (BrokenMarshalJSON) MarshalJSON() ([]byte, error) {
	return nil, errBrokenMarshalJSON
}

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
		case "/broken_marshal":
			return BrokenMarshalJSON{}, nil
		}
		panic("other error")
	})
	// invoke the handler a bunch of times
	fmt.Println(makeRequest(handler, "/value"))
	fmt.Println(makeRequest(handler, "/slice"))
	fmt.Println(makeRequest(handler, "/notfound"))
	fmt.Println(makeRequest(handler, "/other"))
	fmt.Println(makeRequest(handler, "/broken_marshal"))

	// Output: /value returned code 200 with location header "" and body "69\n"
	// /slice returned code 200 with location header "" and body "[\"hello\",42]\n"
	// /notfound returned code 404 with location header "" and body "{\"code\":404,\"status\":\"Not Found\"}"
	// /other returned code 500 with location header "" and body "{\"code\":500,\"status\":\"Internal Server Error\"}"
	// /broken_marshal returned code 500 with location header "" and body ""
}

func TestJSON_LogJSONEncodeError(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		Path       string
		WantCalled bool
	}{
		{Path: "/ok", WantCalled: false},
		{Path: "/broken", WantCalled: true},
	} {
		t.Run(tt.Path, func(t *testing.T) {
			t.Parallel()

			called := false
			handler := content.JSON(func(r *http.Request) (any, error) {
				switch r.URL.Path {
				case "/ok":
					return nil, nil
				case "/broken":
					return BrokenMarshalJSON{}, nil
				}
				panic("never reached")
			})
			handler.LogJSONEncodeError = func(r *http.Request, err error) {
				called = true
			}

			makeRequest(handler, tt.Path)

			if called != tt.WantCalled {
				t.Errorf("want called = %t, got called = %t", tt.WantCalled, called)
			}
		})
	}
}
