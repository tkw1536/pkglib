// Package sema implements semaphores and semaphore-related scheduling
package sema

// New creates a new [Semaphore], guarding a resource of at most the given size.
// The resource is assumed to be entirely available.
//
// A size <= 0 indicates an infinite limit, and all Lock and Unlock calls are no-ops.
//
// Note that if size is statically known to be 1, a Mutex should be used instead.
func New(size int) Semaphore {
	var sema Semaphore
	if size > 0 {
		sema.c = make(chan struct{}, size)
	}
	return sema
}

// Semaphore guards parallel access to a shared resource.
// It should always be created using [New].
//
// The resource can be acquired using a call to [Lock].
// and released using a call to [Unlock].
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
	if s.c == nil {
		return 0
	}
	return len(s.c)
}

// Lock atomically acquires a unit of the guarded resource.
// When the resource is not available, it blocks until such a resource is available.
func (s Semaphore) Lock() {
	if s.c == nil {
		return
	}
	s.c <- struct{}{}
}

// TryLock attempts to atomically acquire the resource without blocking.
// When it succeeds, it returns true, otherwise it returns false.
//
// Calls to [TryLock] never block; they always return immediately.
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
	select {
	case <-s.c:
	default:
		if s.c == nil {
			return
		}
		panic("Semaphore: Unlock without Lock")
	}
}
