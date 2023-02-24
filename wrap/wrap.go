// Package wrap provides facilities to wrap text
package wrap

import (
	"io"
	"strings"
	"sync"
)

var builderPool = &sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// String is a convenience method that creates a new Wrapper, writes s to it, and then returns the written data.
// When s has a trailing newline, also adds a trailing newline to the return value.
func String(length int, s string) string {
	builder := builderPool.Get().(*strings.Builder)
	builder.Reset()
	defer builderPool.Put(builder)

	Write(builder, length, s)
	if strings.HasSuffix(s, "\n") {
		builder.Write(newLine)
	}

	return builder.String()
}

var wrapperPool = &sync.Pool{
	New: func() interface{} {
		return new(Wrapper)
	},
}

// Write is a convenience method that creates a new wrapper and calls the write method on it.
func Write(writer io.Writer, length int, s string) (int, error) {
	// NOTE(twiesing): This method is untested because Wrapper.Write is tested

	wrapper := wrapperPool.Get().(*Wrapper)
	wrapper.Writer = writer
	wrapper.Length = length

	defer func() {
		wrapper.Writer = nil // avoid leaking writer
		wrapperPool.Put(wrapper)
	}()

	return wrapper.WriteString(s)
}
