//spellchecker:words lifetime
package lifetime_test

//spellchecker:words slices strings github pkglib lifetime
import (
	"fmt"
	"slices"
	"strings"

	"github.com/tkw1536/pkglib/lifetime"
)

// Demonstrates the use of the Retrieve function.
// Reuses types from Example B.
func ExampleRetrieve() {
	// Create a new lifetime which registers the same examples as in the previous example.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*House](context)
			lifetime.Place[*Window](context)
		},
	}

	// get the house (as before)
	house := lifetime.Export[*House](lt, struct{}{})

	// magically extract the window!
	window, err := lifetime.Retrieve[*Window, Component](house)
	if err != nil {
		panic(err)
	}

	window.Open()

	// Output: opening the window
}

// Demonstrates the use of the Retrieve function.
// Reuses types from Example E.
func ExampleRetrieveSlice() {
	// Create a new lifetime which registers the same examples as in the previous example.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// retrieve the Wheel (as before).
	wheel := lifetime.Export[*Wheel](lt, struct{}{})

	// magically extract the window!
	colors, err := lifetime.RetrieveSlice[ColorComponent, Component](wheel)
	if err != nil {
		panic(err)
	}

	// sort them according to color!
	slices.SortFunc(colors, func(a, b ColorComponent) int {
		return strings.Compare(a.Color(), b.Color())
	})

	for _, color := range colors {
		fmt.Println(color.Color())
	}

	// Output: green
	// rainbow
	// red
}
