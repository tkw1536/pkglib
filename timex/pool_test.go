//spellchecker:words timex
package timex_test

//spellchecker:words time github pkglib timex
import (
	"fmt"
	"time"

	"github.com/tkw1536/pkglib/timex"
)

func ExampleNewTimer() {
	// take a new timer from the pool
	// and release it again when done
	t := timex.NewTimer()
	defer timex.ReleaseTimer(t)

	// the returned timer is stopped
	// so it will never fire!
	select {
	case <-t.C:
		fmt.Println("timer fired initially")
	case <-time.After(2 * timex.Short):
		fmt.Println("timer did not fire initially")
	}

	// if you reset it, it will fire!
	t.Reset(timex.Short)
	select {
	case <-t.C:
		fmt.Println("timer fired")
	case <-time.After(2 * timex.Short):
		fmt.Println("timer did not fire")
	}

	// Output: timer did not fire initially
	// timer fired
}
