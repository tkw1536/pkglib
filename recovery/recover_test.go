package recovery_test

import (
	"errors"
	"fmt"

	"github.com/tkw1536/pkglib/recovery"
)

func ExampleSafe() {
	// a function that doesn't return an error is invoked normally
	fmt.Println(
		recovery.Safe(func() (int, error) {
			return 42, nil
		}),
	)

	fmt.Println(
		recovery.Safe(func() (int, error) {
			return 0, errors.New("some error")
		}),
	)

	{
		res, err := recovery.Safe(func() (int, error) {
			panic("test panic")
		})

		fmt.Printf("%d %#v\n", res, err)
	}
	// Output: 42 <nil>
	// 0 some error
	// 0 recovery.recovered{/* recover() = "test panic" */}
}
