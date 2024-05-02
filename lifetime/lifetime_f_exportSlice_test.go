//spellchecker:words lifetime
package lifetime_test

//spellchecker:words slices strings github pkglib lifetime
import (
	"fmt"
	"slices"
	"strings"

	"github.com/tkw1536/pkglib/lifetime"
)

// Demonstrates the use of the ExportSlice function.
// Reuses types from Example E.
func ExampleLifetime_fExportSlice() {
	// Create a new lifetime which registers the same examples as in the previous example.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// this time retrieve multiple components using the ExportSlice function.
	colors := lifetime.ExportSlice[ColorComponent](lt, struct{}{})

	// sort them according to their color
	slices.SortFunc(colors, func(a, b ColorComponent) int {
		return strings.Compare(a.Color(), b.Color())
	})

	// and print their colors
	for _, c := range colors {
		fmt.Println(c.Color())
	}

	// Output: green
	// rainbow
	// red
}
