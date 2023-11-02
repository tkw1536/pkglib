package lifetime_test

import (
	"fmt"

	"github.com/tkw1536/pkglib/lifetime"
)

// ShadyFigure is a component that depends on a secret.
type ShadyFigure struct {
	dependencies struct {
		Secret *Secret
	}
}

func (sf *ShadyFigure) SayTheSecret() {
	fmt.Println("The secret is:", sf.dependencies.Secret.value)
}

func (*ShadyFigure) isComponent() {}

// Secret is a component that holds a secret value.
type Secret struct {
	value int
}

func (*Secret) isComponent() {}

// Demonstrates the use of an InitParam.
func ExampleLifetime_initParam() {
	// Create a new lifetime that registers all the components from above.
	lt := &lifetime.Lifetime[Component, int]{
		Register: func(context *lifetime.RegisterContext[Component, int]) {
			// register the ShadyFigure as normal
			lifetime.Place[*ShadyFigure](context)

			// register the Secret component and store the secret value.
			lifetime.Register[*Secret](context, func(s *Secret, secret int) {
				s.value = secret
			})
		},
	}

	// get the shady figure, passing in the secret to be set
	shady := lifetime.Export[*ShadyFigure](lt, 108)
	shady.SayTheSecret()

	// Output: The secret is: 108
}
