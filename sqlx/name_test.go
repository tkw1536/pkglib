package sqlx

import (
	"testing"
)

func TestIsSafeDatabaseLiteral(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"system", false},
		{"", false},
		{"example", true},
		{"@thing", true},
		{"#thing", true},
		{"_thing", true},
		{"$thing", false},
		{"123thing", false},
		{"nothing124$else", true},
		{"hello$world", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafeDatabaseLiteral(tt.name); got != tt.want {
				t.Errorf("IsSafeDatabaseLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSafeDatabaseSingleQuote(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"hello world", true},
		{"hello`world", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafeDatabaseSingleQuote(tt.name); got != tt.want {
				t.Errorf("IsSafeDatabaseSingleQuote() = %v, want %v", got, tt.want)
			}
		})
	}
}
