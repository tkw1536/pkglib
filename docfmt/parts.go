//spellchecker:words docfmt
package docfmt

//spellchecker:words strconv unicode
import (
	"strconv"
	"unicode"
)

// SplitParts splits a message into parts for validation.
//
// Message parts are delimited by either ':' or '.'.
// Each part may contain quoted strings, as in go syntax.
// Quoted strings are always considered part of the same part.
//
// Separators are considered part of the preceding part.
// Every part, with the exception of the last part, will have a string separator.
// The empty message consists of a single empty part.
//
// Any message fulfills the invariant:
//
//	message == strings.Join(SplitParts(message), "")
func SplitParts(message string) (parts []string) {
	return splitString([]rune(message), isPartSeparator, false)
}

// isPartSeparator checks if r is a part separator.
func isPartSeparator(r rune) bool {
	return r == ':' || r == '.'
}

// SplitWords splits a single part into different words and a possibly trailing separator.
//
// Words are delimited by space characters.
// Each part may contain quoted strings, as in go syntax.
// Quoted strings are always considered part of the same word.
//
// Separators are considered part of the preceding word.
// Every word, with the exception of the last word, will end in a non-empty sequence of whitespace characters
// The empty part consists of a single empty word.
//
// Any part fulfills the invariant:
//
//	words, sep := SplitWords(part)
//	part == strings.Join(SplitParts(words), "") + sep
func SplitWords(part string) (words []string, sep string) {
	if part == "" {
		return
	}

	runes := []rune(part)

	// trim of the separator (if any)
	last := runes[len(runes)-1]
	if isPartSeparator(last) {
		sep = string(last)
		runes = runes[:len(runes)-1]
	}

	// split into words
	return splitString(runes, unicode.IsSpace, true), sep
}

// splitString splits runes into strings delimited by runes that contain isDelimited.
// Each part can be grouped by quoting using golang syntax.
func splitString(runes []rune, isDelimiter func(rune) bool, multiDelim bool) (parts []string) {
	var start int // start of the current part

	for index := 0; index < len(runes); index++ {
		switch runes[index] {
		case '`', '"', '\'':
			prefix, err := strconv.QuotedPrefix(string(runes[index:]))
			if err == nil { // syntax error, treat as a normal character
				index += len(prefix) - 1 // skip over the rest of the string
			}
		default:
			if isDelimiter(runes[index]) { // ending the current part
				// allow multiple sequential delimiters to form a single ending
				// so gobble delimiters until they no longer match
				if multiDelim {
					index++
					for index < len(runes) && isDelimiter(runes[index]) {
						index++
					}
					index--
				}
				parts = append(parts, string(runes[start:index+1]))
				start = index + 1
			}
		}
	}

	parts = append(parts, string(runes[start:])) // append the last part!

	return
}
