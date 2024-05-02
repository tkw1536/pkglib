//go:build windows

// Package umaskfree provides file system functionality that ignore the umask.
// As such it is not safe to use otherwise equivalent functions provided by the standard go library concurrently with this package.
// Users should take care that no other code in their application uses these functions.
//
// On windows, using this package has no effect over the normal functions.
//
//spellchecker:words umaskfree
package umaskfree

// mask is the global mask lock
var m mask

// mask does nothing
type mask struct{}

func (mask *mask) Lock() {}

func (mask *mask) Unlock() {}
