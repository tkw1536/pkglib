package sema

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"slices"
)

type Thing uint64

func (t *Thing) OK() error {
	fmt.Printf("OK(%d)\n", *t)
	return nil
}

var errTest = errors.New("test error")

func (t *Thing) Fail() error {
	fmt.Printf("Fail(%d)\n", *t)
	return errTest
}

func (t *Thing) Close() {
	fmt.Printf("Close(%d)\n", *t)
}

func ExamplePool() {
	var counter atomic.Uint64
	p := Pool[*Thing]{
		Limit: 1, // at most one item in the pool
		New: func() *Thing {
			// create a new thing (and print creating it)
			id := counter.Add(1)
			fmt.Printf("New(%d)\n", id)
			return ((*Thing)(&id))
		},
		Discard: (*Thing).Close,
	}
	defer p.Close()

	// the first time an item from the pool is requested, it is created using New()
	p.Use((*Thing).OK)

	// calling it again, re-uses it
	p.Use((*Thing).OK)

	// failing causes it to be destroyed
	p.Use((*Thing).Fail)

	// and calling it again re-creates another one
	p.Use((*Thing).OK)

	// Output: New(1)
	// OK(1)
	// OK(1)
	// Fail(1)
	// Close(1)
	// New(2)
	// OK(2)
	// Close(2)
}

func ExamplePool_Limit() {
	var counter atomic.Uint64

	N := 10
	M := 1000

	var destroyedM sync.Mutex
	var destroyed []uint64

	p := Pool[uint64]{
		Limit: 10, // at most one item in the pool
		New: func() uint64 {
			return counter.Add(1)
		},
		Discard: func(u uint64) {
			destroyedM.Lock()
			defer destroyedM.Unlock()
			destroyed = append(destroyed, u)
		},
	}

	// fill the pool up with N items
	var wg sync.WaitGroup
	wg.Add(N)
	done := make(chan struct{})

	for i := 0; i < N; i++ {
		go p.Use(func(u uint64) error {
			wg.Done() // tell the outer loop an item has been created
			<-done    // do not return the item to the pool until all have been created
			return nil
		})
	}
	wg.Wait()
	close(done)

	// use the items a bunch of times
	wg.Add(M)
	for i := 0; i < M; i++ {
		go func() {
			defer wg.Done()
			p.Use(func(u uint64) error { return nil })
		}()
	}
	wg.Wait()

	// destroy all of them
	// (this will record the items destroyed)
	p.Close()

	// we never had more than 10 items!
	slices.Sort(destroyed)
	fmt.Println(destroyed)

	// Output: [1 2 3 4 5 6 7 8 9 10]

}
