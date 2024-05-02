// Package timex manages various timer-related functions.
//
//spellchecker:words timex
package timex

//spellchecker:words context time
import (
	"context"
	"fmt"
	"time"
)

func ExampleStopTimer() {

	// create a new timer and stop it
	t := time.NewTimer(short)
	StopTimer(t)

	// wait twice the timer amount to make sure it did not fire!
	select {
	case <-t.C:
		fmt.Println("timer fired")
	case <-time.After(2 * short):
		fmt.Println("timer did not fire")
	}

	// Output: timer did not fire
}

func ExampleStopTimer_fired() {

	// create a new timer to fire pretty much immediately
	t := time.NewTimer(time.Nanosecond)

	// wait for a bit, then stop the timer
	time.Sleep(short)
	StopTimer(t)

	// check if the timer fired
	fired := false
	select {
	case <-t.C:
		fired = true
	case <-time.After(short):
		fired = false
	}
	fmt.Println(fired)

	// Output: false
}

func ExampleTickContext() {
	// create a new context
	ctx, cancel := context.WithCancel(context.Background())

	ticker := TickContext(ctx, short)

	// have a couple ticks, each time the channel should be open
	{
		_, ok := <-ticker
		fmt.Println("tick(1)", ok)
	}

	{
		_, ok := <-ticker
		fmt.Println("tick(2)", ok)
	}

	// cancel the context, don't tick any further
	cancel()

	// the channel is now closed
	{
		_, ok := <-ticker
		fmt.Println("tick(3)", ok)
	}

	// Output: tick(1) true
	// tick(2) true
	// tick(3) false
}

func ExampleTickUntilFunc() {
	var counter int

	// keep a counter, and stop when it reaches 3!
	_ = TickUntilFunc(func(t time.Time) bool {
		counter++
		fmt.Printf("tick(%d)\n", counter)
		return counter == 3
	}, context.Background(), short)

	// Output: tick(1)
	// tick(2)
	// tick(3)
}
