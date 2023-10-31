package lifetime_test

import (
	"fmt"
	"slices"

	"github.com/tkw1536/pkglib/lifetime"
)

// Component is the type of Components used by the lifetime.
// Every component must implement it with a pointer receiver.
type Component interface {
	// Here we have a simple method that indicates a struct implements component.
	// In a real-life scenario there would likely be further methods.
	isComponent()
}

// Simple is a simple component that does nothing.
type Simple struct{}

func (*Simple) isComponent() {}

// WithInjectField has a dependency automatically registered using the inject field
type WithInjectField struct {
	simple *Simple `inject:"true"`
}

func (wif *WithInjectField) MustHaveSimpleSet() {
	if wif == nil || wif.simple == nil {
		panic("does not have simple set")
	}
}
func (*WithInjectField) isComponent() {}

// Odd is a component that can check if a number is odd
type Odd struct {
	Setup bool // set during initialization

	dependencies struct {
		Even *Even
	}
}

// IsOdd checks if a number is odd by potentially recursing into Even
func (odd *Odd) IsOdd(value uint) bool {
	if !odd.Setup {
		panic("Odd.Setup unset")
	}

	if value == 0 {
		return true
	}
	return !odd.dependencies.Even.IsEven(value - 1)
}

func (Odd) isComponent() {}

// Even is a component that can check if a number is even
type Even struct {
	Setup bool // set during initialization

	dependencies struct {
		Odd *Odd
	}
}

func (Even) isComponent() {}

// IsEven checks if a number is even by potentially recursing into Odd
func (even *Even) IsEven(value uint) bool {
	if !even.Setup {
		panic("Even.Setup unset")
	}
	if value == 0 {
		return true
	}
	return !even.dependencies.Odd.IsOdd(value - 1)
}

// Color is something that can return a color string
type Color interface {
	Component
	Color() string
}

type Red struct{}

func (Red) isComponent()  {}
func (Red) Color() string { return "red" }

type Green struct{}

func (Green) isComponent()  {}
func (Green) Color() string { return "green" }

type Wheel struct {
	dependencies struct {
		Colors []Color
	}
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

// ExampleLifetime demonstrates the simple use of a lifetime for managing a set of components
func ExampleLifetime() {
	// Create a new lifetime that registers all the components from above.
	lt := &lifetime.Lifetime[Component, int]{
		Register: func(context *lifetime.RegisterContext[Component, int]) {

			// the even and odd components need to be initialized
			lifetime.Register(context, func(even *Even, _ int) { even.Setup = true })
			lifetime.Register(context, func(odd *Odd, _ int) { odd.Setup = true })

			// all the other components do not require initialization
			lifetime.Place[*Red](context)
			lifetime.Place[*Green](context)
			lifetime.Place[*Wheel](context)
			lifetime.Place[*WithInjectField](context)
			lifetime.Place[*Simple](context)
		},
	}

	// you can retrieve specific components from the lifetime
	// where all the other components have been set
	{
		even := lifetime.Export[*Even](lt, 0)

		fmt.Printf("12 is even: %t\n", even.IsEven(12))
		fmt.Printf("13 is even: %t\n", even.IsEven(13))
	}

	// slices of components are also appropriately filled
	{
		wheel := lifetime.Export[*Wheel](lt, 0)
		fmt.Printf("colors are: %v\n", wheel.Colors())
	}

	// "auto" fields can also be filled automatically
	{
		wac := lifetime.Export[*WithInjectField](lt, 0)
		wac.MustHaveSimpleSet()
	}

	// Output: 12 is even: true
	// 13 is even: false
	// colors are: [green red]
}
