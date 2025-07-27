//spellchecker:words lifetime
package lifetime_test

//spellchecker:words pkglib lifetime
import (
	"fmt"

	"go.tkw01536.de/pkglib/lifetime"
)

// Box is a component that depends on a secret.
type Box struct {
	dependencies struct {
		Secret *Secret
	}
}

func (*Box) isComponent() {}

// RevealSecret reveals the secret.
func (b *Box) RevealSecret() {
	fmt.Println("The secret is:", b.dependencies.Secret.value)
}

// Secret is a component that holds a value.
type Secret struct {
	value int
}

func (*Secret) isComponent() {}

// Demonstrates the use of an InitParam within a Lifetime.
func ExampleLifetime_dInitParam() {
	// Declare a lifetime, same as before.
	// Notice that we now pass an additional parameter of type int.
	lt := &lifetime.Lifetime[Component, int]{
		Register: func(context *lifetime.Registry[Component, int]) {
			lifetime.Place[*Box](context)

			// We use the Register function to perform additional initialization.
			// Here, we store the passed parameter into the secret value.
			lifetime.Register(context, func(s *Secret, secret int) {
				s.value = secret
			})
		},
	}

	// Again retrieve the a component using the Export function.
	// This time, pass the additional parameter into all calls to Init.
	box := lifetime.Export[*Box](lt, 108)
	box.RevealSecret()

	// Output: The secret is: 108
}
