// Package sema implements semaphores and semaphore-related scheduling
//
//spellchecker:words sema
package sema

// New creates a new [Semaphore], guarding a resource of at most the given size.
// The resource is assumed to be entirely available.
//
// A size <= 0 indicates an infinite limit, and all Lock and Unlock calls are no-ops.
// A size == 1 indicates a [sync.Mutex] should be used instead.
func New(size int) Semaphore {
	var sema Semaphore
	if size > 0 {
		sema.c = make(chan struct{}, size)
	}
	return sema
}

// Semaphore guards concurrent access to a shared resource.
// The resource can be acquired using [Semaphore.Lock] and released using [Semaphore.Unlock].
//
// A Semaphore is typically created using [New], the zero value implements all operations as no-ops.
type Semaphore struct {
	// if the limit is infinite, the channel is nil;
	// else it is a buffered channel.
	//
	// to acquire a resource, it is written into the underlying buffer;
	// to release a resource, it is read from the buffer;
	c chan struct{}
}

// Len returns the maximum size of the resource guarded by this semaphore, or 0 if said limit is infinite.
func (s Semaphore) Len() int {
	return len(s.c)
}

// Lock acquires the guarded resource.
// If the resource is unavailable, the calling goroutine
// blocks until it is.
func (s Semaphore) Lock() {
	if s.c == nil {
		return
	}
	s.c <- struct{}{}
}

// TryLock tries to acquire s and reports whether it succeeded.
// Calls never block, and always return immediately.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of semaphores.
func (s Semaphore) TryLock() bool {
	if s.c == nil {
		return true
	}

	select {
	case s.c <- struct{}{}:
		return true
	default:
		return false
	}
}

// Unlock releases one unit of the resource that has been previously acquired.
// Calls to Unlock never block.
//
// Calls to Unlock without an acquired resource are a programming error and may block forever.
func (s Semaphore) Unlock() {
	if s.c == nil {
		return
	}

	select {
	case <-s.c:
	default:
		panic("Semaphore: Unlock without Lock")
	}
}
