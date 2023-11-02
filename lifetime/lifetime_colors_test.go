package lifetime_test

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tkw1536/pkglib/lifetime"
)

// ColorComponent is component that has a Color method
type ColorComponent interface {
	Component
	Color() string
}

// Red is a ColorComponent
type Red struct{}

func (Red) isComponent()  {}
func (Red) Color() string { return "red" }

// Green is a ColorComponent
type Green struct{}

func (Green) isComponent()  {}
func (Green) Color() string { return "green" }

// Wheel can list all ColorComponents.
// It is also a ColorComponent itself.
type Wheel struct {
	dependencies struct {
		Colors []ColorComponent
	}
}

func (Wheel) Color() string {
	return "rainbow"
}

func (Wheel) isComponent() {}

func (wheel *Wheel) Colors() []string {
	colors := make([]string, 0, len(wheel.dependencies.Colors))
	for _, c := range wheel.dependencies.Colors {
		colors = append(colors, c.Color())
	}
	slices.Sort(colors)
	return colors
}

// Demonstrates the use of slices in dependencies.
func ExampleLifetime_slicedependencies() {
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.RegisterContext[Component, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// retrieve the wheel component
	wheel := lifetime.Export[*Wheel](lt, struct{}{})
	fmt.Printf("wheel knows the following colors: %v\n", wheel.Colors())

	// Output: wheel knows the following colors: [green rainbow red]
}

// Demonstrates the use of the ExportSlice function.
func ExampleLifetime_exportSlice() {
	// Create a new lifetime which registers all components
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.RegisterContext[Component, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// export all color components by hand
	colors := lifetime.ExportSlice[ColorComponent](lt, struct{}{})

	// and sort them according to their color
	slices.SortFunc(colors, func(a, b ColorComponent) int {
		return strings.Compare(a.Color(), b.Color())
	})

	for _, c := range colors {
		fmt.Println(c.Color())
	}

	// Output: green
	// rainbow
	// red
}

// Demonstrate the use of an init hook.
func ExampleLifetime_init() {
	var colors []string

	lt := &lifetime.Lifetime[ColorComponent, struct{}]{
		// Init is called for each component after it has been initialized.
		Init: func(c ColorComponent, s struct{}) {
			// after initializing a component, add it's color to the struct
			colors = append(colors, c.Color())
		},
		Register: func(context *lifetime.RegisterContext[ColorComponent, struct{}]) {
			lifetime.Place[*Wheel](context)
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
		},
	}

	// force initialization of the lifetime
	_ = lifetime.Export[*Wheel](lt, struct{}{})

	// sort the colors and print them out
	slices.Sort(colors)
	fmt.Printf("Initialized colors: %s\n", colors)

	// Output: Initialized colors: [green rainbow red]
}
