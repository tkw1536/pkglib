//spellchecker:words traversal
package traversal_test

//spellchecker:words errors github pkglib traversal
import (
	"errors"
	"fmt"

	"github.com/tkw1536/pkglib/traversal"
)

func ExampleEmpty() {
	thing := traversal.Empty[any](nil)
	fmt.Println(thing.Next())
	fmt.Println(thing.Err())

	// Output: false
	// <nil>
}

func ExampleEmpty_error() {
	thing := traversal.Empty[any](errors.New("some error"))
	fmt.Println(thing.Next())
	fmt.Println(thing.Err())

	// Output: false
	// some error
}
