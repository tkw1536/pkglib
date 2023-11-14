package lifetime_test

import (
	"fmt"
	"slices"

	"github.com/tkw1536/pkglib/lifetime"
)

// Demonstrates the use of the Init hook.
// Reuses types from Example E.
func ExampleLifetime_hInit() {
	// for purposes of this example keep a list of initialized colors.
	var colors []string

	lt := &lifetime.Lifetime[ColorComponent, struct{}]{
		// Init is called for each component after it has been initialized.
		// Here we just store that the color has been seen.
		Init: func(c ColorComponent, s struct{}) {
			colors = append(colors, c.Color())
		},
		Register: func(context *lifetime.Registry[ColorComponent, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// All initializes and retrieves all components.
	// This will call the Init function above.
	lt.All(struct{}{})

	// Sort the colors we saw and print them out
	slices.Sort(colors)
	fmt.Printf("Initialized colors: %s\n", colors)

	// Output: Initialized colors: [green rainbow red]
}
