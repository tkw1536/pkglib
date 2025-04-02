//spellchecker:words recovery
package recovery_test

//spellchecker:words errors github pkglib recovery
import (
	"errors"
	"fmt"

	"github.com/tkw1536/pkglib/recovery"
)

var errSomething = errors.New("something")

func ExampleSafe() {
	// a function that doesn't return an error is invoked normally
	fmt.Println(
		recovery.Safe(func() (int, error) {
			return 42, nil
		}),
	)

	fmt.Println(
		recovery.Safe(func() (int, error) {
			return 0, errSomething
		}),
	)

	{
		res, err := recovery.Safe(func() (int, error) {
			panic("test panic")
		})

		fmt.Printf("%d %#v\n", res, err)
	}
	// Output: 42 <nil>
	// 0 something
	// 0 recovery.recovered{/* recover() = "test panic" */}
}
