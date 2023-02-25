// Package nobufio provides non-buffered io operations.
package nobufio

import (
	"io"
	"unicode/utf8"
)

// ReadRune reads the next rune from r.
// It does not read from reader beyond the rune.
//
// It returns the rune being read, and its' size in bytes.
// If no rune can be read, it returns an error.
//
// See [io.RuneReader].
func ReadRune(reader io.Reader) (r rune, size int, err error) {
	// try to directly read the rune
	if rreader, ok := reader.(io.RuneReader); ok {
		return rreader.ReadRune()
	}

	var runeBuffer []byte
	for !utf8.FullRune(runeBuffer) {
		// expand the rune buffer
		runeBuffer = append(runeBuffer, 0)

		// read the next byte into it into or bail out!
		if _, err = reader.Read(runeBuffer[size:]); err != nil {
			return
		}
		size++
	}

	// decode the rune!
	r, _ = utf8.DecodeRune(runeBuffer)
	return r, size, nil
}
