package lifetime_test

import (
	"fmt"

	"github.com/tkw1536/pkglib/lifetime"
)

// Odd is a component that can check if a number is odd.
type Odd struct {
	dependencies struct {
		Even *Even
	}
}

func (Odd) isComponent() {}

func (odd *Odd) IsOdd(value uint) bool {
	if value == 0 {
		return true
	}
	return !odd.dependencies.Even.IsEven(value - 1)
}

// Even is a component that can check if a number is even.
type Even struct {
	dependencies struct {
		Odd *Odd
	}
}

func (Even) isComponent() {}

func (even *Even) IsEven(value uint) bool {
	if value == 0 {
		return true
	}
	return !even.dependencies.Odd.IsOdd(value - 1)
}

// Demonstrates the use of mutually dependent components.
func ExampleLifetime_mutual() {
	lt := &lifetime.Lifetime[Component, struct{}]{
		Register: func(context *lifetime.RegisterContext[Component, struct{}]) {
			lifetime.Place[*Even](context)
			lifetime.Place[*Odd](context)
		},
	}

	// retrieve the even component, and do some checking
	even := lifetime.Export[*Even](lt, struct{}{})
	fmt.Printf("42 is even: %t\n", even.IsEven(42))
	fmt.Printf("69 is even: %t\n", even.IsEven(69))

	// Output: 42 is even: true
	// 69 is even: false
}
