package docfmt

import (
	"reflect"
	"testing"
)

// partTests are tests where an entire part is checked against the Validate() or Check() function.
// These are outlined because they are reuse between parts
var partTests = []struct {
	name      string
	input     string
	wantError []ValidationResult
}{
	// failed checks
	{"empty part", "hello::world", []ValidationResult{{PartIndex: 1, WordIndex: 0, Part: ":", Word: "", Kind: WordIsEmpty}}},
	{"may not have extra spaces", "hello: world  ", []ValidationResult{{PartIndex: 1, WordIndex: 2, Part: " world  ", Word: "", Kind: WordIsEmpty}}},
	{"may not start with upper case", "Hello World", []ValidationResult{{PartIndex: 0, WordIndex: 0, Part: "Hello World", Word: "Hello", Kind: WordForbiddenRune}, {PartIndex: 0, WordIndex: 1, Part: "Hello World", Word: "World", Kind: WordForbiddenRune}}},
	{"may not start with upper case (2)", "HeLLo World", []ValidationResult{{PartIndex: 0, WordIndex: 0, Part: "HeLLo World", Word: "HeLLo", Kind: WordForbiddenRune}, {PartIndex: 0, WordIndex: 1, Part: "HeLLo World", Word: "World", Kind: WordForbiddenRune}}},
	{"may not have an invalid ending spaces", "hello\tworld", []ValidationResult{{PartIndex: 0, WordIndex: 0, Part: "hello\tworld", Word: "hello\t", Kind: WordInvalidEnd}}},

	// passed checks
	{"empty string passes", "", nil},
	{"string with multiple parts passes", "something: something else: something else again", nil},
	{"string with entire uppercase word passes", "SOMETHING: something else: something else again", nil},

	// ... word tests added by init ...
}

// wordTests are tests that can be used for testing either a single word, or as a part containing a single word
// When asPart is false, they cannot be used as a part test.
var wordTests = []struct {
	input   string
	want    ValidationKind
	comment string
	asPart  bool
}{

	{"nothing", ValidationOK, "only letters", true},
	{"(nothing)", ValidationOK, "leading and trailing bracket is ok", true},
	{"(nothing", ValidationOK, "leading only bracket", true},
	{"nothing)", ValidationOK, "trailing only bracket", true},
	{"nothing,", ValidationOK, "trailing comma", true},
	{"hello-world", ValidationOK, "letters and dashes", true},
	{"1234", ValidationOK, "only numbers", true},
	{"1", ValidationOK, "only numbers", true},
	{"URL", ValidationOK, "only capitals", true},
	{"URLs", ValidationOK, "capitals, followed by 's'", true},
	{"%s", ValidationOK, "leading % allowed", true},
	{"`something'", ValidationOK, "`' delimited word is allowed", true},
	{`"hello W0rld-"`, ValidationOK, " \"-quoted word is ok", true},
	{"`hello W0rld-`", ValidationOK, " `-quoted word is ok", true},
	{"`hello W0rld-`,", ValidationOK, " `-quoted word with suffix is ok", true},

	{"_", WordForbiddenRune, "dash not allowed", true},
	{"a1", WordForbiddenRune, "mixed letter numbers not allowed", true},
	{`"hello world"extra stuff`, WordIncorrectQuote, "may not have extra content after quote", false},
	{`"unclosed`, WordIncorrectQuote, "unclosed quote", true},
	{"hello world", WordForbiddenRune, "spaces not allowed", false}, // exluded from part test because it won't be seen as a part
	{"Hello", WordForbiddenRune, "capital letter not allowed", true},
	{"hello--world", WordNoSequentialDashes, "non sequential dashes not allowed", true},
	{"-hello", WordNoOutsideDashes, "leading dash not allowed", true},
	{"hello-", WordNoOutsideDashes, "trailing dash not allowed", true},
	{"hello(", WordForbiddenRune, "inside bracket not allowed", true},
	{"noth,ing", WordForbiddenRune, "non-trailing comma not allowed", true},
	{"som%thing", WordForbiddenRune, "% in the middle of word not allowed", true},
	{"a%", WordForbiddenRune, "trailing % not allowed", true},
}

// add the word tests to the part tests
func init() {
	for _, wt := range wordTests {
		if !wt.asPart {
			continue
		}

		var wantError []ValidationResult
		if wt.want != ValidationOK {
			wantError = []ValidationResult{
				{
					PartIndex: 0,
					Part:      wt.input,
					WordIndex: 0,
					Word:      wt.input,
					Kind:      wt.want,
				},
			}
		}
		_ = wantError

		partTests = append(partTests, struct {
			name      string
			input     string
			wantError []ValidationResult
		}{
			name:      wt.comment,
			input:     wt.input,
			wantError: wantError,
		})
	}
}

func TestValidate(t *testing.T) {
	for _, tt := range partTests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := Validate(tt.input)

			if !reflect.DeepEqual(gotErr, tt.wantError) {
				t.Errorf("Validate() error = %#v, want = %#v", gotErr, tt.wantError)
			}
		})
	}
}

func Test_validateWord(t *testing.T) {
	for _, tt := range wordTests {
		t.Run(tt.comment, func(t *testing.T) {
			if got := validateWord(tt.input); got != tt.want {
				t.Errorf("validateWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
