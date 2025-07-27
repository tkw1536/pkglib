//spellchecker:words text
package text_test

//spellchecker:words reflect testing pkglib text
import (
	"reflect"
	"testing"

	"go.tkw01536.de/pkglib/text"
)

func TestSplitter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		haystack string
		chars    string
		want     []string
	}{
		{
			"hello world",
			" ",
			[]string{"hello", "world"},
		},
		{
			"hello:world;how:is:it:going",
			":;",
			[]string{"hello", "world", "how", "is", "it", "going"},
		},
		{
			":::;;;",
			":;",
			[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.haystack, func(t *testing.T) {
			t.Parallel()
			if got := text.Splitter(tt.chars)(tt.haystack); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Splitter() = %v, want %v", got, tt.want)
			}
		})
	}
}
