// Package text provides functions similar to strings.Join, but based on writers as opposed to strings
//
//spellchecker:words text
package text

import (
	"io"

	"go.tkw01536.de/pkglib/sequence"
)

//spellchecker:words nolint wrapcheck

// Join writes the elements of elem into writer, separated by sep.
// Returns the number of runes written and a nil error.
//
// It is like [strings.Join], but writes into a writer instead of allocating a [strings.Builder].
//
//nolint:wrapcheck // so that it matches strings.Join better
func Join(w io.Writer, elems []string, sep string) (n int, err error) {
	// this function has been adapted from strings.Join

	switch len(elems) {
	case 0:
		return
	case 1:
		return io.WriteString(w, elems[0])
	}

	// count how many elements we'll have to write
	if grower, ok := w.(interface {
		Grow(length int)
	}); ok {
		n := len(sep) * (len(elems) - 1)
		for _, elem := range elems {
			n += len(elem)
		}
		grower.Grow(n)
	}

	sw := sequence.Writer{Writer: w}

	_, _ = sw.WriteString(elems[0])
	for _, s := range elems[1:] {
		_, _ = sw.WriteString(sep)
		_, _ = sw.WriteString(s)
	}

	return sw.Sum()
}

// RepeatJoin writes s, followed by (count -1) instances of sep + s into w.
// It returns the number of runes written and a nil error.
//
// When count <= 0, no instances of s or sep are written into count.
func RepeatJoin(w io.Writer, s, sep string, count int) (n int, err error) {
	if count <= 0 {
		return
	}

	if grower, ok := w.(interface {
		Grow(length int)
	}); ok {
		n = len(s)*count + len(sep)*(count-1)
		grower.Grow(n)
	}

	sw := sequence.Writer{Writer: w}

	_, _ = sw.WriteString(s)
	for range count - 1 {
		_, _ = sw.WriteString(sep + s)
	}

	//nolint:wrapcheck // explicitly return the underlying error
	return sw.Sum()
}

// Repeat writes count instances of s into w.
// It returns the number of runes written and a nil error.
// When count would cause an overflow, calls panic().
//
// It is similar to strings.Repeat, but writes into an existing builder without allocating a new one.
//
// When s is empty or count <= 0, no instances of s are written.
func Repeat(w io.Writer, s string, count int) (n int, err error) {
	// this function has been adapted from strings.Repeat
	// with the only significant change being that we track an additional offset in builder!

	if count <= 0 || s == "" {
		return
	}

	if len(s)*count/count != len(s) {
		panic("Repeat: Repeat count causes overflow")
	}

	// grow the buffer by the overall number of bytes needed
	if grower, ok := w.(interface {
		Grow(length int)
	}); ok {
		n = len(s) * count
		grower.Grow(n)
	}

	// do the writing!
	sw := sequence.Writer{Writer: w}
	for range count {
		_, _ = sw.WriteString(s)
	}

	//nolint:wrapcheck // explicitly return the underlying error
	return sw.Sum()
}
