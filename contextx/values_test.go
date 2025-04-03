//spellchecker:words contextx
package contextx_test

//spellchecker:words context github pkglib contextx
import (
	"context"
	"fmt"

	"github.com/tkw1536/pkglib/contextx"
)

type customInt int

var (
	oneContextKey   = customInt(1)
	twoContextKey   = customInt(2)
	threeContextKey = customInt(3)
	fourContextKey  = customInt(4)
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

func ExampleWithValuesOf() {
	// create a primary context with 'hello' values.
	// Note that we only fill '1' and '2' here.
	primary := contextx.WithValues(context.Background(), map[any]any{
		oneContextKey: "hello earth",
		twoContextKey: "hello mars",
	})

	// create a secondary context with 'bye' values
	// note that we only fill '2' and '3' here.
	secondary := contextx.WithValues(context.Background(), map[any]any{
		twoContextKey:   "bye mars",
		threeContextKey: "bye venus",
	})

	// now creates a derived context that overrides the values of primary with secondary.
	derived := contextx.WithValuesOf(primary, secondary)

	// found only in primary
	fmt.Println(derived.Value(oneContextKey))

	// found in both, the secondary overrides
	fmt.Println(derived.Value(twoContextKey))

	// found only in secondary
	fmt.Println(derived.Value(threeContextKey))

	// found in neither
	fmt.Println(derived.Value(fourContextKey))

	// Output: hello earth
	// bye mars
	// bye venus
	// <nil>
}
