package trigger

import (
	"sync"
)

// Trigger represents an object for which either multiple non-exclusive holds, or a single exclusive hold can be acquired.
// Holds must not be acquired recursively within the same call.
//
// A trigger is that to be in "held" state if at least one hold has been acquired.
// Otherwise it is said to be in "unheld" state.
// Whenever it switches from unheld to held state, OnAcquire is called.
// Whenever it switches from held to unheld state, OnRelease is called.
//
// The zero value is ready for holds to be acquired; it represents a trigger in "unheld" state.
type Trigger struct {
	l       sync.RWMutex // held by any "client"
	s       sync.Mutex   // held when changing state
	counter int64

	OnAcquire func(exclusive bool)
	OnRelease func(exclusive bool)
}

// Lock blocks until it can acquire a non-exclusive hold on this Trigger.
// Each call to Lock may be undone by a call to Unlock.
//
// If this call switches the trigger from a non-held to a held state, OnAcquire(false) is called unless OnAcquire is nil.
// Any concurrent calls to Lock or Unlock will not return until this OnAcquire call has returned.
// If OnAquire panics, the trigger is considered locked.
func (trigger *Trigger) Lock() {
	trigger.l.RLock()

	trigger.s.Lock()
	defer trigger.s.Unlock()

	// increase the counter, and run onAcquire
	trigger.counter++
	if trigger.counter == 1 && trigger.OnAcquire != nil {
		trigger.OnAcquire(false)
	}
}

// Unlock releases a non-exclusive hold on this Trigger.
// There should be an equal amount of calls to Unlock as to Lock.
// Making more unlock calls results in a panic.
//
// If this call switches the trigger from a held to a non-held state, OnRelease(false) is called unless OnRelease is nil.
// Any concurrent calls to Unlock or Lock will not return until this OnRelease call has returned.
// If OnRelease panics, the trigger is considered unlocked.
func (trigger *Trigger) Unlock() {
	(func() {
		trigger.s.Lock()
		defer trigger.s.Unlock()

		if trigger.counter == 0 {
			panic("Trigger.Unlock: Too many unlock calls")
		}

		trigger.counter--
		if trigger.counter == 0 && trigger.OnRelease != nil {
			trigger.OnRelease(false)
		}

	})()

	trigger.l.RUnlock()
}

// XLock blocks until an exclusive lock on this trigger can be acquired.
// It then calls trigger.OnAcquire(true) (unless OnAcquire is nil).
//
// No other calls to Lock or XLock succeed before XUnlock is called.
// If OnAcquire panics, the lock is considered acquired.
func (trigger *Trigger) XLock() {
	trigger.l.Lock()

	if trigger.OnAcquire != nil {
		trigger.OnAcquire(true)
	}
}

// XUnlock calls trigger.OnRelease(true), and then releases an exclusive hold on trigger.
//
// If onRelease is nil, it is not called.
// If OnRelease panics, the trigger is considered unlocked.
func (trigger *Trigger) XUnlock() {
	defer trigger.l.Unlock()

	if trigger.OnRelease != nil {
		trigger.OnRelease(true)
	}
}
