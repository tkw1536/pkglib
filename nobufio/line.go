//spellchecker:words nobufio
package nobufio

//spellchecker:words strings
import (
	"io"
	"strings"
)

//spellchecker:words errorlint nolint wrapcheck

// ReadLine reads the current line from the provided reader.
// It does not use a buffer, and does not read beyond the end of line marker.
//
// A line is considered to end when one of the following is encountered: '\r\n', '\n' or EOF or '\r' followed by EOF.
// Note that only a '\r' alone is not considered an end-of-line.
//
// The returned line never contains the end-of-line markers, such as '\n' or '\r\n'.
// A line may be empty, however when only EOF is read, returns "", EOF.
func ReadLine(reader io.Reader) (value string, err error) {
	var builder strings.Builder // buffer for the string to construct
	var lastR bool              // delay writing a '\r', in case it is followed by an '\n'
	var readSomething bool
	for {
		// read the next valid rune
		r, _, err := ReadRune(reader)
		if err == io.EOF { //nolint:errorlint // io.EOF
			break // at EOF, we are done!
		}
		readSomething = true
		if err != nil { // unknown reading error => bail out
			return "", err
		}
		if r == '\n' { // \n or \r\n
			break
		}

		if lastR {
			// flag is set, but we didn't encounter a '\n' or EOF.
			// so we need to write it back to the buffer
			if _, err := builder.WriteRune('\r'); err != nil {
				panic("never reached")
			}
			lastR = false
		}
		if r == '\r' {
			lastR = true
			continue
		}

		// store it to the builder
		if _, err := builder.WriteRune(r); err != nil {
			panic("never reached")
		}
	}

	// if we didn't read anything, return EOF!
	if !readSomething {
		return "", io.EOF
	}

	// make it a string
	return builder.String(), nil
}
