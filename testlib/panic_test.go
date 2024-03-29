// Package testlib contains functions useful for testing
package testlib

import (
	"fmt"
)

// DoesPanic behavior for a panicking function
func ExampleDoesPanic_panic() {
	didPanic, recovered := DoesPanic(func() {
		panic("some error message")
	})
	fmt.Printf("didPanic = %t\n", didPanic)
	fmt.Printf("recover() = %v\n", recovered)
	// Output: didPanic = true
	// recover() = some error message
}

// DoesPanic behavior for a function that calls panic(nil)
func ExampleDoesPanic_nil() {
	didPanic, _ := DoesPanic(func() {
		panic(nil)
	})
	fmt.Printf("didPanic = %t\n", didPanic)
	// Output: didPanic = true
}

// DoesPanic behavior for a function that does not panic
func ExampleDoesPanic_normal() {
	didPanic, _ := DoesPanic(func() {
		/* do nothing */
	})
	fmt.Printf("didPanic = %t\n", didPanic)
	// Output: didPanic = false
}
