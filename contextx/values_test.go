//spellchecker:words contextx
package contextx_test

//spellchecker:words context pkglib contextx
import (
	"context"
	"fmt"

	"go.tkw01536.de/pkglib/contextx"
)

type customInt int

var (
	oneContextKey   = customInt(1)
	twoContextKey   = customInt(2)
	threeContextKey = customInt(3)
)

func ExampleWithValues() {
	// create a background context without any values
	original := context.Background()

	// add two values to it!
	derived := contextx.WithValues(original, map[any]any{
		oneContextKey: "hello earth",
		twoContextKey: "hello mars",
	})

	// we just set these above
	fmt.Println(derived.Value(oneContextKey))
	fmt.Println(derived.Value(twoContextKey))

	// this context key has nothing associated with it
	fmt.Println(derived.Value(threeContextKey))

	// Output: hello earth
	// hello mars
	// <nil>
}
