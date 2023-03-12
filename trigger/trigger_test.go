package trigger

import (
	"fmt"
	"testing"
)

var largeNumber = 1000

func TestTrigger_Lock(t *testing.T) {
	var trigger Trigger
	trigger.OnAcquire = func(exclusive bool) {
		fmt.Println("lock acquired exclusive=", exclusive)
	}
	trigger.OnRelease = func(released bool) {
		fmt.Println("lock acquired exclusive=", released)
	}

	// acquire and release a lot of locks
	for i := 0; i < largeNumber; i++ {
		trigger.Lock()
	}
	for i := 0; i < largeNumber; i++ {
		trigger.Unlock()
	}

	// take an exclusive lock
	trigger.XLock()

	done := make(chan struct{})
	go func() {
		defer close(done)

		// acquire and release a lot of locks
		// this won't succeed until the exclusive lock is released!
		for i := 0; i < largeNumber; i++ {
			trigger.Lock()
		}
		for i := 0; i < largeNumber; i++ {
			trigger.Unlock()
		}
	}()

	trigger.XUnlock()

	<-done

	// Output: lock acquired exclusive=false
	// lock released exclusive=false
	// lock acquired exclusive=true
	// lock released exclusive=true
	// lock acquired exclusive=true
	// lock released exclusive=false
}
