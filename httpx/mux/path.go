package mux

import (
	"path"
)

// NormalizePath normalizes a path sent to a webserver.
//
// - Any "." and ".." in path are removed lexicographically
// - If any path does not start with "/", it is prepended to the string
// - If any path does not end with "/", it is appended.
// - Paths that start with ".." or "." are turned in paths starting at the root.
func NormalizePath(value string) string {
	if value == "" {
		return "/"
	}

	// ensure that we have a leading slash
	if value[0] != '/' {
		value = "/" + value
	}

	// clean the value
	value = path.Clean(value)

	// The cleaned path ends in a slash if and only if it is the root path.
	// Add it back otherwise.
	if value == "/" {
		return value
	}
	return value + "/"
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
