// Package password allows generating random passwords
//
//spellchecker:words password
package password

//spellchecker:words math strings crypto rand crand
import (
	"io"
	"math/big"
	"strings"

	crand "crypto/rand"
)

// Charset represents a set of runes to include in a password.
// An empty Charset is equivalent to [DefaultCharSet].
type Charset string

func (c Charset) Contains(r rune) bool {
	if c == "" {
		return DefaultCharSet.Contains(r)
	}

	return strings.ContainsRune(string(c), r)
}

// ContainsOnly checks if password contains only runes from this Charset.
func (c Charset) ContainsOnly(password string) bool {
	for _, r := range password {
		if !c.Contains(r) {
			return false
		}
	}
	return true
}

// DefaultCharset represents the default Charset to use.
const DefaultCharSet Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// rand is used as a source for randomness.
func Generate(rand io.Reader, length int, charset Charset) (string, error) {
	if length < 0 {
		panic("length < 0")
	}

	// determine the charset to use
	if charset == "" {
		charset = DefaultCharSet
	}

	// extract the possible runes and their count
	runes := []rune(charset)
	runeCount := big.NewInt(int64(len(runes)))

	// create a buffer to write the string to!
	var password strings.Builder
	password.Grow(length)

	for range length {
		// grab a random bIndex!
		bIndex, err := crand.Int(rand, runeCount)
		if err != nil {
			return "", err
		}

		// and use that index!
		index := int(bIndex.Int64())
		if _, err := password.WriteRune(runes[index]); err != nil {
			return "", err
		}
	}

	// return the password!
	return password.String(), nil
}
