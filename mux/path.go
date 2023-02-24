package mux

import (
	"path"
)

// TESTME

// NormalizePath normalizes the provided path.
// This passes it to "path".Clean, and then ensures there is both a leading and trailing slash.
func NormalizePath(value string) string {
	value = path.Clean(value)
	if value != "/" {
		value = value + "/"
	}
	return value
}

// parentSegment returns the parent segment of the provided path
// it assumes that normalizePath has been called on value.
func parentSegment(value string) string {
	if value == "" || value == "/" {
		return "/"
	}
	parent := path.Dir(value[:len(value)-1])
	if parent != "/" {
		parent = parent + "/"
	}
	return parent
}
