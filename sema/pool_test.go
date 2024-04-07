package sema_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tkw1536/pkglib/sema"
)

func TestPool_Limit(t *testing.T) {

	for _, tt := range []struct {
		Name              string
		Limit, Iterations int
	}{
		{"single item pool", 1, 1_000},
		{"small pool", 10, 1_000},
		{"medium pool", 100, 10_000},
		{"huge pool", 1000, 100_000},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			var (
				createdCount atomic.Int64 // count of created items

				destroyedCount  atomic.Int64 // count of destroyed items
				destroyedValues sync.Map     // values having been destroyed
			)

			p := sema.Pool[int64]{
				// setup the limit
				Limit: tt.Limit,

				// creating simply increases the count
				New: func() int64 {
					return createdCount.Add(1)
				},

				// Discard
				Discard: func(u int64) {
					destroyedCount.Add(1)
					destroyedValues.Store(u, struct{}{})
				},
			}

			// fill the pool up with N items
			var wg sync.WaitGroup
			wg.Add(tt.Limit)
			done := make(chan struct{})

			for range tt.Limit {
				go p.Use(func(u int64) error {
					wg.Done() // tell the outer loop an item has been created
					<-done    // do not return the item to the pool until all have been created
					return nil
				})
			}
			wg.Wait()
			close(done)

			if created := int(createdCount.Load()); created != tt.Limit {
				t.Errorf("created %d items(s), but expected %d", created, tt.Limit)
			}

			// use the items a bunch of times
			wg.Add(tt.Iterations)
			for range tt.Iterations {
				go func() {
					defer wg.Done()
					p.Use(func(u int64) error { return nil })
				}()
			}
			wg.Wait()

			// destroy all of them (will record destruction)
			p.Close()

			// check that the right amount of items was destroyed
			if dc := int(destroyedCount.Load()); dc != tt.Limit {
				t.Errorf("destroyed %d items(s), but expected %d", dc, tt.Limit)
			}

			// check that each item was destroyed
			for i := range int64(tt.Limit) {
				i := i + 1
				_, ok := destroyedValues.Load(i)
				if !ok {
					t.Errorf("item %d was not destroyed", i)
				}
			}
		})
	}
}
