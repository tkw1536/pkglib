// Package text provides functions similar to strings.Join, but based on writers as opposed to strings
//
//spellchecker:words text
package text

import (
	"io"
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

	// figure out what to use for WriteString
	writeString := stringWriter(w)

	// write the first string
	{
		m, err := writeString(elems[0])
		n += m
		if err != nil {
			return n, err
		}
	}

	for _, s := range elems[1:] {
		// write a separator
		{
			m, err := writeString(sep)
			n += m
			if err != nil {
				return n, err
			}
		}

		// write the next string
		{
			m, err := writeString(s)
			n += m
			if err != nil {
				return n, err
			}
		}
	}

	return n, nil
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

	writeString := stringWriter(w)

	m, err := writeString(s)
	if err != nil {
		return m, err
	}

	if n, err := repeat(writeString, sep+s, count-1); err != nil {
		return m + n, err
	}

	return n, nil
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

	// do the actual repeat
	if n, err := repeat(stringWriter(w), s, count); err != nil {
		return n, err
	}

	return n, nil
}

// stringWriter returns a function that does io.WriteString(w, ...)
// It is used for init time branching.
func stringWriter(w io.Writer) func(string) (int, error) {
	if sw, ok := w.(io.StringWriter); ok {
		return sw.WriteString
	}

	return func(s string) (int, error) {
		return w.Write([]byte(s))
	}
}

// only compute the number of bytes written if something goes wrong.
func repeat(w func(string) (int, error), s string, count int) (int, error) {
	// NOTE: This function exists to save having to repeatedly call
	// io.WriteString; which always rechecks if the passed type fulfils the interface.
	for i := range count {
		if m, err := w(s); err != nil {
			return len(s)*i + m, err
		}
	}
	return 0, nil
}
