// Package testlib provides utilities for testing
package testlib

// DoesPanic runs f and checks if it panicked or not.
// When f does panic, returns the recovered value.
func DoesPanic(f func()) (panicked bool, recovered interface{}) {

	// In principle this function could just return the value of recover.
	// However that wouldn't allow to tell the difference between f calling panic(nil) and not panicking at all.
	// TODO(go.21): Just do the recover and check if it is non-nil (because it has been replaced).

	defer func() {
		recovered = recover()
	}()

	panicked = true
	f()
	panicked = false

	return
}
