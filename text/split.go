package text

import "strings"

// Splitter returns a function that acts like [strings.Split],
// except that it splits on any rune contained in chars.
func Splitter(chars string) func(string) []string {
	// create a set of runes to be included in the split
	runes := []rune(chars)
	if len(runes) == 1 {
		// handle common case of just a single char
		return func(s string) []string { return strings.Split(s, chars) }
	}

	isSplit := make(map[rune]struct{}, len(runes))
	for _, r := range runes {
		isSplit[r] = struct{}{}
	}

	// return a function that creates the split map
	return func(s string) []string {
		return strings.FieldsFunc(s, func(r rune) bool {
			_, ok := isSplit[r]
			return ok
		})
	}
}
