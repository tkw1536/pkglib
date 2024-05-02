//spellchecker:words text
package text

//spellchecker:words reflect testing
import (
	"reflect"
	"testing"
)

func TestSplitter(t *testing.T) {
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
			if got := Splitter(tt.chars)(tt.haystack); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Splitter() = %v, want %v", got, tt.want)
			}
		})
	}
}
