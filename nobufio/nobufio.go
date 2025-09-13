// Package nobufio provides non-buffered io operations.
//
//spellchecker:words nobufio
package nobufio

//spellchecker:words unicode
import (
	"io"
	"sync"
	"unicode/utf8"
)

//spellchecker:words nolint wrapcheck

// ReadRune reads the next rune from r.
// If r is a [io.RuneReader], it will be used directly.
// Otherwise it reads exactly as many bytes from reader as needed to decode the full rune.
//
// It returns the rune being read, and its' size in bytes.
// If no rune can be read, it returns an error.
func ReadRune(reader io.Reader) (r rune, size int, err error) {
	// try to directly read the rune
	if reader, ok := reader.(io.RuneReader); ok {
		return reader.ReadRune() //nolint:wrapcheck // directly use RuneReader
	}

	return readRuneSlow(reader)
}

// runeBufferPool contains []byte of size [utf8.UTFMax].
var runeBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, utf8.UTFMax)
		return &buf
	},
}

// readRuneSlow reads the next rune from reader.
// It does not read more bytes than needed to decode the full rune.
func readRuneSlow(reader io.Reader) (r rune, size int, err error) {
	runeBuffer := *(runeBufferPool.Get().(*[]byte))
	runeBuffer = runeBuffer[:0]
	defer runeBufferPool.Put(&runeBuffer)

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
