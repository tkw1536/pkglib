//spellchecker:words docfmt
package docfmt_test

//spellchecker:words reflect testing github pkglib docfmt
import (
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/docfmt"
)

func TestSplitParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantParts []string
	}{
		{"", []string{""}},
		{"hello world", []string{"hello world"}},
		{"hello: world", []string{"hello:", " world"}},
		{"hello: world:", []string{"hello:", " world:", ""}},
		{`hello ": world": you are awesome`, []string{"hello \": world\":", " you are awesome"}},
		{`hello.world`, []string{"hello.", "world"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			if gotParts := docfmt.SplitParts(tt.input); !reflect.DeepEqual(gotParts, tt.wantParts) {
				t.Errorf("SplitParts() = %#v, want %#v", gotParts, tt.wantParts)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantWords []string
		wantSep   string
	}{
		{"", []string(nil), ""},
		{"hello world.", []string{"hello ", "world"}, "."},
		{"multiple         whitespace", []string{"multiple         ", "whitespace"}, ""},
		{"trailing whitespace ", []string{"trailing ", "whitespace ", ""}, ""},
		{"multiple trailing whitespace       ", []string{"multiple ", "trailing ", "whitespace       ", ""}, ""},
		{"word ` with a quote` something", []string{"word ", "` with a quote` ", "something"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			gotWords, gotSep := docfmt.SplitWords(tt.input)
			if !reflect.DeepEqual(gotWords, tt.wantWords) {
				t.Errorf("SplitWords() gotWords = %#v, want %#v", gotWords, tt.wantWords)
			}
			if gotSep != tt.wantSep {
				t.Errorf("SplitWords() gotSep = %v, want %v", gotSep, tt.wantSep)
			}
		})
	}
}
