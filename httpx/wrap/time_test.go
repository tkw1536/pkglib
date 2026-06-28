//spellchecker:words wrap
package wrap_test

//spellchecker:words context encoding json http httptest time pkglib httpx wrap
import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"go.tkw01536.de/pkglib/httpx/wrap"
)

func ExampleTime() {
	// delay used during this example
	var delay = 100 * time.Millisecond

	// Create a new HandlerFunc that sleeps for the delay.
	handler := wrap.Time(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)         // sleep for a bit
		took := wrap.TimeSince(r) // record how long it took

		_ = json.MarshalWrite(w, took.Nanoseconds())
	}))

	// create a new request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	if err != nil {
		panic(err)
	}

	// serve the http request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		fmt.Println("Expected http.StatusOK")
	}

	// decode the amount of time taken from the request
	var tookNanoseconds int64
	_ = json.UnmarshalRead(rr.Result().Body, &tookNanoseconds)

	took := time.Duration(tookNanoseconds)
	if took >= delay {
		fmt.Println("Handler returned correct delay")
	} else {
		fmt.Printf("got tok = %v < %v = delay", took, delay)
	}

	// Output: Handler returned correct delay
}
