package docfmt

import (
	"reflect"
	"testing"
)

func TestSplitParts(t *testing.T) {
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
			if gotParts := SplitParts(tt.input); !reflect.DeepEqual(gotParts, tt.wantParts) {
				t.Errorf("SplitParts() = %#v, want %#v", gotParts, tt.wantParts)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
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
			gotWords, gotSep := SplitWords(tt.input)
			if !reflect.DeepEqual(gotWords, tt.wantWords) {
				t.Errorf("SplitWords() gotWords = %#v, want %#v", gotWords, tt.wantWords)
			}
			if gotSep != tt.wantSep {
				t.Errorf("SplitWords() gotSep = %v, want %v", gotSep, tt.wantSep)
			}
		})
	}
}
