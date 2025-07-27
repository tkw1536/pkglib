//spellchecker:words sema
package sema_test

//spellchecker:words errors sync atomic time pkglib sema
import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"go.tkw01536.de/pkglib/sema"
)

func ExampleSchedule() {
	var counter atomic.Uint64

	// because we have a parallelism of 1, we run exactly in order!
	_ = sema.Schedule(func(i uint64) error {
		counter.Add(1)
		return nil
	}, 1000, sema.Concurrency{
		Limit: 0,
		Force: false,
	})

	fmt.Print(counter.Load())
	// Output: 1000
}

func ExampleSchedule_order() {
	// because we have a parallelism of 1, we run exactly in order!
	_ = sema.Schedule(func(i uint64) error {
		fmt.Print(i, ";")
		return nil
	}, 4, sema.Concurrency{
		Limit: 1,
		Force: false,
	})

	// Output: 0;1;2;3;
}

var errFirstFunction = errors.New("first function error")

func ExampleSchedule_error() {
	err := sema.Schedule(func(i uint64) error {
		// the first invocation produces an error and returns immediately!
		if i == 0 {
			return errFirstFunction
		}

		// concurrently with the first invocation, we have at most one other
		// so give the first function some time to produce an error
		time.Sleep(100 * time.Millisecond)

		// the third and fourth invocations should never be called
		// since by the time the first function finishes
		// the second one has already produced an error
		if i > 1 {
			panic("never reached")
		}
		return nil
	}, 4, sema.Concurrency{
		Limit: 2,
		Force: false,
	})
	fmt.Println(err)

	// Output: first function error
}

func ExampleSchedule_force() {
	var counter atomic.Uint64

	err := sema.Schedule(func(i uint64) error {
		// count the number of invocations
		counter.Add(1)

		// the first function returns an error
		// but because of force = True, the execution is not aborted
		if i == 0 {
			return errFirstFunction
		}

		// ... work ...
		time.Sleep(50 * time.Millisecond)
		return nil
	}, 10, sema.Concurrency{
		Limit: 2,
		Force: true,
	})

	fmt.Println(err)
	fmt.Println(counter.Load(), "workers called")

	// Output: first function error
	// 10 workers called
}
