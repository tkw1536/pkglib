// Package wrap provides facilities to wrap text
package wrap

// spellchecker:words twiesing

import (
	"io"
	"strings"
)

// String is a convenience method that creates a new Wrapper, writes s to it, and then returns the written data.
// When s has a trailing newline, also adds a trailing newline to the return value.
//
// Deprecated: Do not use package wrap.
func String(length int, s string) string {
	var builder strings.Builder

	Write(&builder, length, s)
	if strings.HasSuffix(s, "\n") {
		builder.Write(newLine)
	}

	return builder.String()
}

// Write is a convenience method that creates a new wrapper and calls the write method on it.
//
// Deprecated: Do not use package wrap.
func Write(writer io.Writer, length int, s string) (int, error) {
	// NOTE(twiesing): This method is untested because Wrapper.Write is tested

	var wrapper Wrapper
	wrapper.Writer = writer
	wrapper.Length = length

	return wrapper.WriteString(s)
}
