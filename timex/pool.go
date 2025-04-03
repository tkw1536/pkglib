//spellchecker:words timex
package timex

//spellchecker:words sync time
import (
	"sync"
	"time"
)

// Short is a short time.Duration to use for various initializations.
const Short = 100 * time.Millisecond

var tPool = sync.Pool{
	New: func() any {
		timer := time.NewTimer(Short)
		StopTimer(timer)
		return timer
	},
}

// NewTimer returns an unused timer from an internal timer pool.
// The timer is guaranteed to be initialized and stopped.
// The timer will not have been created with AfterFunc.
//
// Before using the timer a call to timer.Reset() should be made.
// The caller is furthermore encouraged to return the timer to the pool using [ReleaseTimer] once it is no longer needed.
func NewTimer() *time.Timer {
	return tPool.Get().(*time.Timer)
}

// ReleaseTimer stops t and returns it to the internal pool of timers.
// t should not have been created with AfterFunc.
func ReleaseTimer(t *time.Timer) {
	StopTimer(t)
	tPool.Put(t)
}

// StopTimer stops the given timer and drains the underlying channel.
// The timer must have been initialized.
//
// This prevents it from firing, until a call to Reset() is made.
// If the timer is not running, StopTimer does nothing.
func StopTimer(t *time.Timer) {
	t.Stop()

	select {
	case <-t.C:
	default:
	}
}
