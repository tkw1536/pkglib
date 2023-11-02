package lifetime_test

import (
	"fmt"

	"github.com/tkw1536/pkglib/lifetime"
)

// Component is the type of Components used by the lifetime.
// Every component must implement it with a pointer receiver.
type Component interface {
	// Here we have a simple method that indicates a struct implements component.
	// In a real-life scenario there would likely be further methods.
	isComponent()
}

// Company is a component that depends on the CEO
type Company struct {
	// explicitly declare a dependencies struct.
	// it must be called "dependencies", and contain references to other components.
	dependencies struct {
		CEO *CEO
	}
}

func (*Company) isComponent() {}

func (s *Company) SayHello() {
	s.dependencies.CEO.SayHello()
	fmt.Println("Hello from the company")
}

// CEO is a component that can say hello
type CEO struct{}

func (*CEO) isComponent() {}

func (CEO) SayHello() {
	fmt.Println("Hello from the CEO")
}

// ExampleLifetime demonstrates the basic use of a lifetime.
func ExampleLifetime() {
	// Create a new lifetime that registers all the components from above.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.RegisterContext[Component, struct{}]) {
			lifetime.Place[*Company](context)
			lifetime.Place[*CEO](context)
		},
	}

	// export the company component
	// which will automatically create both the company and CEO components.
	company := lifetime.Export[*Company](lt, struct{}{})
	company.SayHello()

	// Output: Hello from the CEO
	//Hello from the company
}
