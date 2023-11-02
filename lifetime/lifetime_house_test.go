package lifetime_test

import (
	"fmt"

	"github.com/tkw1536/pkglib/lifetime"
)

// House is a component with a public window
type House struct {
	// Window is directly injected (outside of a dependencies struct)
	// By marking it with the `inject:"true"` tag
	Window *Window `inject:"true"`
}

func (*House) isComponent() {}

// Window is a component that has an open method
type Window struct {
}

func (window *Window) isComponent() {}

func (window *Window) Open() {
	fmt.Println("opening the window")
}

// Demonstrates the use of an inject tag.
func ExampleLifetime_injectTag() {
	// Create a new lifetime that registers all the components from above.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.RegisterContext[Component, struct{}]) {
			lifetime.Place[*House](context)
			lifetime.Place[*Window](context)
		},
	}

	// get the house and open the window
	house := lifetime.Export[*House](lt, struct{}{})
	house.Window.Open()

	// Output: opening the window
}
