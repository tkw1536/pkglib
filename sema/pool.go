package sema

import (
	"sync"

	"slices"

	"github.com/tkw1536/pkglib/lazy"
)

// Pool holds a finite set of lazily created objects.
type Pool[V any] struct {
	// New creates a new item in the pool.
	New func() V

	// Discard is called before an item is (permanently) remove from the pool.
	// Unlike a finalizer, Discard is guaranteed to be called.
	Discard func(V)

	// Limit is the maximum number of objects in the pool.
	// Limit <= 0 means there is no limit on the number of objects in the pool.
	// When Use requests an item, and the limit is already reached,
	// Use will block until an item becomes available.
	Limit int
	s     lazy.Lazy[sync.Locker] // implements limit

	l     sync.Mutex // lock protects items
	items []V        // items holds the current items in the pool

}

// Use borrows an object from the pool, passes it to f, and then returns it to the pool.
// Use blocks until f has returned.
// If f returns an error (or f panics) the returned object is discarded, and a new object is created once needed.
func (pool *Pool[V]) Use(f func(V) error) error {
	// ensure that at most limit calls are active at the same time.
	// this implements the limit on the items in the pool.
	sema := pool.s.Get(func() sync.Locker { return New(pool.Limit) })
	sema.Lock()
	defer sema.Unlock()

	// get or create a new item in the pool
	var entry V
	{
		pool.l.Lock()

		if len(pool.items) > 0 {
			last := len(pool.items) - 1
			entry = pool.items[last]
			pool.items = slices.Clip(pool.items[:last])
		} else {
			entry = pool.New()
		}

		pool.l.Unlock()
	}

	// run the actual function
	ok, err := func() (ok bool, err error) {
		defer func() { _ = recover() }() // silently ignore errors

		ok = false
		err = f(entry)
		ok = err == nil
		return
	}()

	// return the item to the pool (if everything went fine)
	{
		if ok {
			pool.l.Lock()
			pool.items = append(pool.items, entry)
			pool.l.Unlock()
		} else {
			pool.Discard(entry)
		}
	}

	// return the error
	return err
}

// Close discards all objects currently in the pool.
// Note that any active Use calls are not waited for.
func (pool *Pool[V]) Close() {
	pool.l.Lock()
	defer pool.l.Unlock()

	// call discard for all items
	for _, item := range pool.items {
		pool.Discard(item)
	}
	pool.items = nil
}
