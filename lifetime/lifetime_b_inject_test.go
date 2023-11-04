package lifetime_test

import (
	"fmt"

	"github.com/tkw1536/pkglib/lifetime"
)

// House is a component with a public window.
type House struct {
	// Window is directly injected (outside of a dependencies struct).
	// This is achieved with the `inject:"true"` tag.
	Window *Window `inject:"true"`
}

func (*House) isComponent() {}

// Window is a component that has an open method.
type Window struct {
}

func (window *Window) isComponent() {}

// Open opens the window.
func (window *Window) Open() {
	fmt.Println("opening the window")
}

// Demonstrates the use of an inject that to declare dependencies.
func ExampleLifetime_bInjectTag() {
	// Same as before, register all components.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*House](context)
			lifetime.Place[*Window](context)
		},
	}

	// we can now retrieve the house component.
	// The window is set automatically.
	house := lifetime.Export[*House](lt, struct{}{})
	house.Window.Open()

	// Output: opening the window
}
