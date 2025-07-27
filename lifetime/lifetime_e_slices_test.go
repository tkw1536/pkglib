//spellchecker:words lifetime
package lifetime_test

//spellchecker:words slices pkglib lifetime
import (
	"fmt"
	"slices"

	"go.tkw01536.de/pkglib/lifetime"
)

// ColorComponent a subtype of component that implements the color method.
type ColorComponent interface {
	Component
	Color() string
}

// Declare two color components Red and Green.

type Red struct{}

func (Red) isComponent()  {}
func (Red) Color() string { return "red" }

type Green struct{}

func (Green) isComponent()  {}
func (Green) Color() string { return "green" }

// Declare the wheel component.
type Wheel struct {
	dependencies struct {
		// Wheel depends on all ColorComponents.
		// Like other components, these are also initialized automatically.
		Colors []ColorComponent
	}
}

func (*Wheel) isComponent() {}

// Wheel itself is also a special color component.
func (*Wheel) Color() string {
	return "rainbow"
}

// Colors returns the list of known colors in alphabetical order.
func (wheel *Wheel) Colors() []string {
	// retrieve all the colors
	colors := make([]string, 0, len(wheel.dependencies.Colors))
	for _, c := range wheel.dependencies.Colors {
		colors = append(colors, c.Color())
	}

	// since there is not a defined order for components, we sort them!
	slices.Sort(colors)
	return colors
}

// Demonstrates the use of slices in dependencies.
func ExampleLifetime_eSlices() {
	// Register components as normal.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// retrieve the Wheel component and use it as expected.
	wheel := lifetime.Export[*Wheel](lt, struct{}{})
	fmt.Printf("wheel knows the following colors: %v\n", wheel.Colors())

	// Output: wheel knows the following colors: [green rainbow red]
}
