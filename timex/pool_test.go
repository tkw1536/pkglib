package timex

import (
	"fmt"
	"time"
)

func ExampleNewTimer() {

	// take a new timer from the pool
	// and release it again when done
	t := NewTimer()
	defer ReleaseTimer(t)

	// the returned timer is stopped
	// so it will never fire!
	select {
	case <-t.C:
		fmt.Println("timer fired initially")
	case <-time.After(2 * short):
		fmt.Println("timer did not fire initially")
	}

	// if you reset it, it will fire!
	t.Reset(short)
	select {
	case <-t.C:
		fmt.Println("timer fired")
	case <-time.After(2 * short):
		fmt.Println("timer did not fire")
	}

	// Output: timer did not fire initially
	// timer fired
}
