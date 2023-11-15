package wrap_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/tkw1536/pkglib/httpx/wrap"
)

func ExampleTime() {

	// delay used during this example
	var delay = 50 * time.Millisecond

	// Create a new HandlerFunc that sleeps for the delay.
	handler := wrap.Time(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)         // sleep for a bit
		took := wrap.TimeSince(r) // record how long it took

		json.NewEncoder(w).Encode(took)
	}))

	// create a new request
	req, err := http.NewRequest(http.MethodGet, "/", nil)
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
	var took time.Duration
	json.NewDecoder(rr.Result().Body).Decode(&took)

	if took >= delay {
		fmt.Println("Handler returned correct delay")
	}

	// Output: Handler returned correct delay
}
