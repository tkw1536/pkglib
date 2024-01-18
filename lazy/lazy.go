package lazy

import (
	"sync"
)

// Lazy holds a lazily initialized value of T.
// A non-zero lazy must not be copied after first use.
//
// Lazy should be used where a single value is initialized and reset at different call sites with undefined order and values.
// A value initialized in one place and read many times should use [sync.OnceValue] instead.
type Lazy[T any] struct {
	m     sync.Mutex // m protects the value of this lazy
	done  bool       // is a value stored?
	value T          // the held value
}

// Get returns the value associated with this Lazy.
//
// If no other call to Get has started or completed an initialization, calls init to initialize the value.
// A nil init function indicates to store the zero value of T.
// If an initialization has been previously completed, the previously stored value is returned.
//
// If init panics, the initialization is considered to be completed.
// Future calls to Get() do not invoke init, and the zero value of T is returned.
//
// Get may safely be called concurrently.
func (lazy *Lazy[T]) Get(init func() T) T {
	if lazy == nil {
		panic("attempt to access (*Lazy[...])(nil)")
	}

	lazy.m.Lock()
	defer lazy.m.Unlock()

	// value is not yet initialized
	if !lazy.done {
		lazy.done = true
		if init != nil {
			lazy.value = init()
		}
	}

	// and return the value!
	return lazy.value
}

// Set atomically sets the value of this lazy.
// Any previously set value will be overwritten.
// Future calls to [Get] will not invoke init.
//
// It may be called concurrently with calls to [Get].
func (lazy *Lazy[T]) Set(value T) {
	if lazy == nil {
		panic("attempt to access (*Lazy[...])(nil)")
	}

	lazy.m.Lock()
	defer lazy.m.Unlock()

	// we store the value now!
	lazy.done = true
	lazy.value = value
}
