//spellchecker:words sema
package sema_test

//spellchecker:words errors sync atomic github pkglib sema
import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/tkw1536/pkglib/sema"
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
	p := sema.Pool[*Thing]{
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
	_ = p.Use((*Thing).OK)

	// calling it again, re-uses it
	_ = p.Use((*Thing).OK)

	// failing causes it to be destroyed
	_ = p.Use((*Thing).Fail)

	// and calling it again re-creates another one
	_ = p.Use((*Thing).OK)

	// Output: New(1)
	// OK(1)
	// OK(1)
	// Fail(1)
	// Close(1)
	// New(2)
	// OK(2)
	// Close(2)
}
