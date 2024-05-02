// Package docfmt implements formatting and checking of user format strings.
//
// Strings are checked at runtime for proper formatting, or by a (simple) static analyzer.
// Checking is disabled by default, but can be enabled by building with the "doccheck" tag.
// See Check.
//
//spellchecker:words docfmt
package docfmt

//spellchecker:words strings sync unicode github pkglib text
import (
	"strings"
	"sync"
	"unicode"

	"github.com/tkw1536/pkglib/text"
)

var builderPool = &sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

// Format formats message before presenting it to a user.
// It passes the message to AssertValid, which may cause a panic if checking is enabled and the word does not pass the checks.
//
// A message is formatted by splitting a message into parts and words.
// It then capitalizes the first non-whitespace word of each part.
//
// See also SplitParts, SplitWords, Check, Capitalize.
func Format(message string) string {
	AssertValid(message)

	// grab a builder from the pool
	builder := builderPool.Get().(*strings.Builder)
	defer builderPool.Put(builder)
	defer builder.Reset()

	// iterate over the parts
	for _, part := range SplitParts(message) {
		// capitalize the first word!
		words, suffix := SplitWords(part)
		for i, word := range words {
			var ok bool
			if words[i], ok = Capitalize(word); ok {
				break
			}
		}

		// and put it into the builder for rewriting
		_, _ = text.Join(builder, words, "")
		builder.WriteString(suffix)
	}

	return builder.String()
}

// Capitalize capitalizes word that passes validation of individual words.
// A word is capitalized by uppercasing the first non-whitespace rune in the word.
//
// Returns the capitalized word, and a boolean true if capitalization was performed,
// or the unchanged word and false if the word contained only whitespace.
func Capitalize(word string) (result string, ok bool) {
	runes := []rune(word)
	for i, r := range runes {
		if !unicode.IsSpace(r) {
			runes[i] = unicode.ToUpper(r)
			return string(runes), true
		}
	}
	return word, false
}
