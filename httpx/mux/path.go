package mux

import (
	"path"
	"path/filepath"
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

	// ensure we have a leading slash
	if value[0] != '/' {
		value = "/" + value
	}

	// do the cleaning
	value = filepath.Clean(value)

	// add the trailing slash
	// required if and only if the path is not the root directory
	if len(value) == 1 && value == "/" {
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
