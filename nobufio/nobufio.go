// Package nobufio provides non-buffered io operations.
//
//spellchecker:words nobufio
package nobufio

//spellchecker:words unicode
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
	if reader, ok := reader.(io.RuneReader); ok {
		return reader.ReadRune()
	}

	runeBuffer := make([]byte, 0, utf8.MaxRune)
	for !utf8.FullRune(runeBuffer) {
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
