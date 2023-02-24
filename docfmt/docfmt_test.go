// Package docfmt implements formatting and checking of user format strings.
//
// Strings are checked at runtime for proper formatting
// Checking is disabled by default, but can be enabled by building with the "doccheck" tag.
// See Check.
package docfmt

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world.", "Hello world."},
		{"multiple: parts", "Multiple: Parts"},
		{"`mY CapitAlizAtion is not changed because quote`: but mine is", "`mY CapitAlizAtion is not changed because quote`: But mine is"},
		{"i am part 1. i am part 2.", "I am part 1. I am part 2."},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := Format(tt.input); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
