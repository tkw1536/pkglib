//spellchecker:words sqlx
package sqlx_test

//spellchecker:words testing pkglib sqlx
import (
	"testing"

	"go.tkw01536.de/pkglib/sqlx"
)

func TestIsSafeDatabaseLiteral(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			if got := sqlx.IsSafeDatabaseLiteral(tt.name); got != tt.want {
				t.Errorf("IsSafeDatabaseLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSafeDatabaseSingleQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{"hello world", true},
		{"hello`world", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := sqlx.IsSafeDatabaseSingleQuote(tt.name); got != tt.want {
				t.Errorf("IsSafeDatabaseSingleQuote() = %v, want %v", got, tt.want)
			}
		})
	}
}
