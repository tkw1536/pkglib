// Package password allows generating random passwords
//
//spellchecker:words password
package password

//spellchecker:words crypto rand testing
import (
	"crypto/rand"
	"testing"
)

func TestPassword(t *testing.T) {
	N := 1000 // number of runs per test case

	tests := []struct {
		name    string
		length  int
		charset Charset
	}{
		{
			name:    "length 10 default charset",
			length:  10,
			charset: DefaultCharSet,
		},
		{
			name:    "length 20 default charset",
			length:  20,
			charset: DefaultCharSet,
		},
		{
			name:    "length 14 custom charset",
			length:  14,
			charset: "abc%^&",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range N {
				candidate, err := Generate(rand.Reader, tt.length, tt.charset)
				if err != nil {
					t.Error(err)
				}
				if len(candidate) != tt.length {
					t.Error("did not generate password of correct length")
				}
				if !tt.charset.ContainsOnly(candidate) {
					t.Error("did not generate password from the correct charset")
				}
			}
		})
	}
}
