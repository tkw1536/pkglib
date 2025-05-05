// Package timex manages various timer-related functions.
//
//spellchecker:words timex
package timex

//spellchecker:words context time
import (
	"context"
	"time"
)

//spellchecker:words nolint wrapcheck

// TickContext is like [time.Tick], but closes the returned channel once the context closes.
// As such it can be recovered by the garbage collector; see [time.TickContext].
//
// Unlike [time.Tick], immediately sends the current time on the given channel.
func TickContext(c context.Context, d time.Duration) <-chan time.Time {
	if d < 0 {
		return nil
	}

	ticker := make(chan time.Time, 1)
	ticker <- time.Now()
	go func() {
		defer close(ticker)

		timer := NewTimer()
		defer ReleaseTimer(timer)

		for {
			timer.Reset(d)

			select {
			case tick := <-timer.C:
				ticker <- tick
			case <-c.Done():
				return
			}
		}
	}()
	return ticker
}

// TickUntilFunc invokes f every d until either context is closed, or f returns true.
// f is invoked once immediately when the timer starts.
//
// TickUntilFunc blocks until f is no longer invoked.
//
// Returns the error of the context (if any).
func TickUntilFunc(f func(t time.Time) bool, c context.Context, d time.Duration) error {
	context, cancel := context.WithCancel(c)
	defer cancel()

	for t := range TickContext(context, d) {
		if f(t) {
			break
		}
	}
	return c.Err() //nolint:wrapcheck // returns the exact context error
}
