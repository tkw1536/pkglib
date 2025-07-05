//spellchecker:words sema
package sema_test

//spellchecker:words sync atomic testing github pkglib sema
import (
	"sync"
	"sync/atomic"
	"testing"

	"go.tkw01536.de/pkglib/sema"
)

func TestPool_Limit(t *testing.T) {
	t.Parallel()

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
			var createGroup, useGroup sync.WaitGroup
			createGroup.Add(tt.Limit)
			useGroup.Add(tt.Limit)
			done := make(chan struct{})

			for range tt.Limit {
				go func() {
					defer useGroup.Done()
					_ = p.Use(func(u int64) error {
						createGroup.Done() // tell the outer loop an item has been created
						<-done             // do not return the item to the pool until all have been created
						return nil
					})
				}()
			}
			createGroup.Wait()
			close(done)

			if created := int(createdCount.Load()); created != tt.Limit {
				t.Errorf("created %d items(s), but expected %d", created, tt.Limit)
			}

			// use the items a bunch of times
			useGroup.Add(tt.Iterations)
			for range tt.Iterations {
				go func() {
					defer useGroup.Done()
					_ = p.Use(func(u int64) error { return nil })
				}()
			}

			// wait until all the uses have returned
			// this should allow close to destroy all of them
			useGroup.Wait()

			// do the closing, which will record destruction
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
