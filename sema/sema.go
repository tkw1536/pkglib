// Package sema implements semaphores and semaphore-related scheduling
package sema

// NewSemaphore creates a new semaphore with the provided limit.
//
// A semaphore with limit = 1 is equivalent to a *sync.Mutex.
// Semaphores with limit <= 0 implement Lock, TryLock and Unlock as no-ops that return immediately.
func New(limit int) Semaphore {
	var sema Semaphore
	if limit > 0 {
		sema.c = make(chan struct{}, limit)
	}
	return sema
}

// Semaphore guards parallel access to a shared resource.
// It should always be created using New().
//
// The resource can be acquired using a call to Lock()
// and released using a call to Unlock().
// The maximum number of parallel accesses is set by using New().
//
// See also New and sync.Locker.
type Semaphore struct {
	// if the limit is infinite, the channel is nil;
	// else it is a buffered channel.
	//
	// to acquire a resource, it is written into the underlying buffer;
	// to release a resource, it is read from the buffer;
	c chan struct{}
}

// Len returns the length of this semaphore, a.k.a the maximum number of parallel Lock() Unlock() blocks that may be active.
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
// Calls to TryLock() never block; they always return immediately.
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
// Calls to Unlock() never block.
//
// Calls to Unlock() without an acquired resource are a programming error and may block forever.
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
