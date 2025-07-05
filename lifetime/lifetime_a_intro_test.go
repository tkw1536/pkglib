//spellchecker:words lifetime
package lifetime_test

//spellchecker:words github pkglib lifetime
import (
	"fmt"

	"go.tkw01536.de/pkglib/lifetime"
)

// Component is used as a type for components for a lifetime.
type Component interface {
	isComponent()
}

// Company is a Component.
type Company struct {
	// Company declares its' dependencies using an embedded "dependencies" struct.
	// This can simply refer to other components and will be set automatically.
	dependencies struct {
		CEO *CEO
	}
}

// SayHello says hello from the company.
func (c *Company) SayHello() {
	// the CEO dependency will be automatically injected at runtime.
	// so the code can just call methods on it.
	c.dependencies.CEO.SayHello()
	fmt.Println("Hello from the company")
}

func (*Company) isComponent() {}

// CEO is another component.
type CEO struct{}

func (*CEO) isComponent() {}

// SayHello says hello from the CEO.
func (ceo *CEO) SayHello() {
	if ceo == nil {
		panic("nil CEO can't say hello")
	}
	fmt.Println("Hello from the CEO")
}

// Introductory example on how to use a lifetime with two components.
func ExampleLifetime_aIntro() {
	// Create a new lifetime using the Component type
	lt := &lifetime.Lifetime[Component, struct{}]{
		// Register must register all components.
		// In this case we register the company and the CEO.
		Register: func(r *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*Company](r)
			lifetime.Place[*CEO](r)
		},
	}

	// To initialize and use a single component, we make use of the Export function.
	// Here it is invoked to retrieve a Company.
	company := lifetime.Export[*Company](lt, struct{}{})
	company.SayHello()

	// Output: Hello from the CEO
	// Hello from the company
}
