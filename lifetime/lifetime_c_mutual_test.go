//spellchecker:words lifetime
package lifetime_test

//spellchecker:words github pkglib lifetime
import (
	"fmt"

	"github.com/tkw1536/pkglib/lifetime"
)

// Declare two mutually dependent Odd and Even components.

type Odd struct {
	dependencies struct {
		Even *Even
	}
}

type Even struct {
	dependencies struct {
		Odd *Odd
	}
}

func (*Odd) isComponent()  {}
func (*Even) isComponent() {}

// Now declare two functions IsOdd and IsEven on their respective components.
// These make use of each other.

func (odd *Odd) IsOdd(value uint) bool {
	if value == 0 {
		return true
	}
	return !odd.dependencies.Even.IsEven(value - 1)
}

func (even *Even) IsEven(value uint) bool {
	if value == 0 {
		return true
	}
	return !even.dependencies.Odd.IsOdd(value - 1)
}

// Demonstrates that components may be mutually dependent.
func ExampleLifetime_cMutual() {
	// Again register both components.
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.Registry[Component, struct{}]) {
			lifetime.Place[*Even](context)
			lifetime.Place[*Odd](context)
		},
	}

	// retrieve the Even component, the mutual dependencies are set correct.
	even := lifetime.Export[*Even](lt, struct{}{})
	fmt.Printf("42 is even: %t\n", even.IsEven(42))
	fmt.Printf("69 is even: %t\n", even.IsEven(69))

	// Output: 42 is even: true
	// 69 is even: false
}
