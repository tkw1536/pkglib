// Package testlib provides utilities for testing
//
//spellchecker:words testlib
package testlib

// DoesPanic runs f and checks if it panicked or not.
// When f does panic, returns the recovered value.
func DoesPanic(f func()) (panicked bool, recovered interface{}) {
	defer func() {
		recovered = recover()
		panicked = recovered != nil
	}()

	f()

	return
}
