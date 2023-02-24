package docfmt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
)

type ValidationResult struct {
	PartIndex, WordIndex int
	Part, Word           string
	Kind                 ValidationKind
}

func (v ValidationResult) Error() string {
	// NOTE(twiesing): This function is untested because it is used only for developing

	if v.WordIndex == -1 {
		return fmt.Sprintf("part %d %q: %s", v.PartIndex, v.Part, v.Kind)
	}
	return fmt.Sprintf("part %d word %d %q: %s", v.PartIndex, v.WordIndex, v.Word, v.Kind)
}

// ValidationKind represents different types of validationn errors
type ValidationKind string

const (
	ValidationOK           ValidationKind = ""
	PartIsEmpty            ValidationKind = "part is empty"
	WordIsEmpty            ValidationKind = "word is empty"
	WordIncorrectQuote     ValidationKind = "word not quoted correctly"
	WordNoOutsideDashes    ValidationKind = "word has leading or trailing dashes"
	WordNoSequentialDashes ValidationKind = "word contains sequential dashes"
	WordForbiddenRune      ValidationKind = "word may only contain lower case letters and dashes"
	WordInvalidEnd         ValidationKind = "word must end with a single ' ' or '\n'"
)

// Validate validates message and returns all validation errors.
//
// Each message is first split into parts, see SplitParts.
// Then each part is validated as follows:
//   - a part (with the exception of the last part) may not be empty (PartIsEmpty)
//
// Furthermore each part is split into words, see SplitWords.
// Then each word is validated as follows:
//   - no word may be empty (WordIsEmpty)
//   - a word may only contain lower case letter (ForbiddenCharacter)
//   - a word starting with ` and ending with ' is always valid, because it is a quoted word
//   - a word starting with ` and not ending with ' must be a word quoted in go syntax, and may not have any extra content (WordIncorrectQuote)
//   - a word may have a leading '(' or a trailing ')', assuming other rules apply accordingly
//   - a word may contain a trailing comma
//   - a word may only contain capital letters when all runes in it are capital letters, or if the last letter is an s and the rest are capital.
//   - a word may consist of only digits when all runes in it are digits
//   - a word may contain '-'s, but only non-sequential occurences (WordNoSequentialDashes) that are not the first or last letter (WordNoOutsideDashes)
//   - a word that starts with '%' is always valid, because it might be a format string
//   - each word (except for the last word) must end with a space character, that is either " " or "\n" (InvalidEndSpace)
func Validate(message string) (errors []ValidationResult) {
	parts := SplitParts(message)
partloop:
	for pI, part := range parts {
		if pI != len(parts)-1 && part == "" {
			errors = append(errors, ValidationResult{
				PartIndex: pI,
				Part:      part,
				WordIndex: -1,
				Kind:      PartIsEmpty,
			})
			continue partloop
		}

		words, _ := SplitWords(part)
	wordloop:
		for wI, word := range words {
			if word == "" {
				errors = append(errors, ValidationResult{
					PartIndex: pI,
					Part:      part,
					WordIndex: wI,
					Word:      word,
					Kind:      WordIsEmpty,
				})
				continue wordloop
			}

			// every non-last word *must* end with a ' '  or '\n' character
			if wI != len(words)-1 {
				if word[len(word)-1] != ' ' && word[len(word)-1] != '\n' {
					errors = append(errors, ValidationResult{
						PartIndex: pI,
						Part:      part,
						WordIndex: wI,
						Word:      word,
						Kind:      WordInvalidEnd,
					})
				}
				word = word[:len(word)-1]
			}
			word = strings.TrimSpace(word)

			// check word for appropriate letters
			if typ := validateWord(word); typ != ValidationOK {
				errors = append(errors, ValidationResult{
					PartIndex: pI,
					Part:      part,
					WordIndex: wI,
					Word:      word,
					Kind:      typ,
				})
			}

		}
	}
	return errors
}

// IsValidWord checks that word fullfills the rules for a valid word
func validateWord(word string) ValidationKind {
	// NOTE(twiesing): Return the exact validation result
	runes := []rune(word)

	// word contained in `' is allowed
	if len(runes) != 0 && runes[0] == '`' && runes[len(runes)-1] == '\'' {
		return ValidationOK
	}

	// leading '(' allowed
	if len(runes) != 0 && runes[0] == '(' {
		runes = runes[1:]
	}

	// trailing ')' allowed
	if len(runes) != 0 && runes[len(runes)-1] == ')' {
		runes = runes[:len(runes)-1]
	}

	// trailing comma allowed
	if len(runes) != 0 && runes[len(runes)-1] == ',' {
		runes = runes[:len(runes)-1]
	}

	// a word starting with " must be golang quoted
	if len(runes) > 0 && (runes[0] == '"' || runes[0] == '`') {
		word := string(runes)
		prefix, err := strconv.QuotedPrefix(word)
		return check(err == nil && prefix == word, WordIncorrectQuote)
	}

	// format string
	if len(runes) != 0 && runes[0] == '%' {
		return ValidationOK
	}

	// all uppercase, ending with s
	if len(runes) != 0 && runes[len(runes)-1] == 's' {
		if value, ok := allEqual(runes[:len(runes)-1], unicode.IsUpper); value && ok {
			return ValidationOK
		}
	}

	// all uppercase
	if value, ok := allEqual(runes, unicode.IsUpper); value && ok {
		return ValidationOK
	}

	// all digits
	if value, ok := allEqual(runes, unicode.IsDigit); value && ok {
		return ValidationOK
	}

	// check that we don't have sequential '-'s
	if !nonSequential(runes, func(r rune) bool { return r == '-' }) {
		return WordNoSequentialDashes
	}

	// leading and trailing -s not allowed
	if len(runes) != 0 && (runes[0] == '-' || runes[len(runes)-1] == '-') {
		return WordNoOutsideDashes
	}

	// only letters
	return check(
		all(runes, func(r rune) bool {
			return r == '-' || unicode.IsLower(r)
		}, true),
		WordForbiddenRune,
	)
}

// allEqual checks that f returns the same value for every element of s
//
// When this is the case, returns the return value and true.
// If not, returns the zero value of V and false.
//
// When elements is empty, returns the zero value of V.
func allEqual[T any, V comparable](s []T, f func(e T) V) (value V, ok bool) {
	if len(s) < 1 {
		return value, true
	}

	first := f(s[0]) // return value for the first element
	for _, element := range s[1:] {
		if f(element) != first {
			return value, false
		}
	}
	return first, true
}

// all checks that f returns v for all elements of s
func all[T any, V comparable](s []T, f func(e T) V, v V) (ok bool) {
	return slices.IndexFunc(s, func(e T) bool { return f(e) != v }) == -1
}

// nonSequential checks that f does not return true for sequential values of s
func nonSequential[T any](s []T, f func(value T) bool) bool {
	if len(s) <= 1 {
		return true
	}

	last := f(s[0])
	for _, element := range s[1:] {
		next := f(element)
		if last && next == last {
			return false
		}
		last = next
	}
	return true
}

// check returns kind if valid is false else ValidationOK
func check(valid bool, kind ValidationKind) ValidationKind {
	if valid {
		return ValidationOK
	}
	return kind
}
