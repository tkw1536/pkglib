//spellchecker:words docfmt
package docfmt_test

//spellchecker:words reflect testing github pkglib docfmt
import (
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/docfmt"
)

//spellchecker:words noth

// These are outlined because they are reuse between parts.
var partTests = []struct {
	name      string
	input     string
	wantError []docfmt.ValidationResult
}{
	// failed checks
	{"empty part", "hello::world", []docfmt.ValidationResult{{PartIndex: 1, WordIndex: 0, Part: ":", Word: "", Kind: docfmt.WordIsEmpty}}},
	{"may not have extra spaces", "hello: world  ", []docfmt.ValidationResult{{PartIndex: 1, WordIndex: 2, Part: " world  ", Word: "", Kind: docfmt.WordIsEmpty}}},
	{"may not start with upper case", "Hello World", []docfmt.ValidationResult{{PartIndex: 0, WordIndex: 0, Part: "Hello World", Word: "Hello", Kind: docfmt.WordForbiddenRune}, {PartIndex: 0, WordIndex: 1, Part: "Hello World", Word: "World", Kind: docfmt.WordForbiddenRune}}},
	{"may not start with upper case (2)", "HeLLo World", []docfmt.ValidationResult{{PartIndex: 0, WordIndex: 0, Part: "HeLLo World", Word: "HeLLo", Kind: docfmt.WordForbiddenRune}, {PartIndex: 0, WordIndex: 1, Part: "HeLLo World", Word: "World", Kind: docfmt.WordForbiddenRune}}},
	{"may not have an invalid ending spaces", "hello\tworld", []docfmt.ValidationResult{{PartIndex: 0, WordIndex: 0, Part: "hello\tworld", Word: "hello\t", Kind: docfmt.WordInvalidEnd}}},

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
	want    docfmt.ValidationKind
	comment string
	asPart  bool
}{

	{"nothing", docfmt.ValidationOK, "only letters", true},
	{"(nothing)", docfmt.ValidationOK, "leading and trailing bracket is ok", true},
	{"(nothing", docfmt.ValidationOK, "leading only bracket", true},
	{"nothing)", docfmt.ValidationOK, "trailing only bracket", true},
	{"nothing,", docfmt.ValidationOK, "trailing comma", true},
	{"hello-world", docfmt.ValidationOK, "letters and dashes", true},
	{"1234", docfmt.ValidationOK, "only numbers", true},
	{"1", docfmt.ValidationOK, "only numbers", true},
	{"URL", docfmt.ValidationOK, "only capitals", true},
	{"URLs", docfmt.ValidationOK, "capitals, followed by 's'", true},
	{"%s", docfmt.ValidationOK, "leading % allowed", true},
	{"`something'", docfmt.ValidationOK, "`' delimited word is allowed", true},
	{`"hello W0rld-"`, docfmt.ValidationOK, " \"-quoted word is ok", true},
	{"`hello W0rld-`", docfmt.ValidationOK, " `-quoted word is ok", true},
	{"`hello W0rld-`,", docfmt.ValidationOK, " `-quoted word with suffix is ok", true},

	{"_", docfmt.WordForbiddenRune, "dash not allowed", true},
	{"a1", docfmt.WordForbiddenRune, "mixed letter numbers not allowed", true},
	{`"hello world"extra stuff`, docfmt.WordIncorrectQuote, "may not have extra content after quote", false},
	{`"unclosed`, docfmt.WordIncorrectQuote, "unclosed quote", true},
	{"hello world", docfmt.WordForbiddenRune, "spaces not allowed", false}, // excluded from part test because it won't be seen as a part
	{"Hello", docfmt.WordForbiddenRune, "capital letter not allowed", true},
	{"hello--world", docfmt.WordNoSequentialDashes, "non sequential dashes not allowed", true},
	{"-hello", docfmt.WordNoOutsideDashes, "leading dash not allowed", true},
	{"hello-", docfmt.WordNoOutsideDashes, "trailing dash not allowed", true},
	{"hello(", docfmt.WordForbiddenRune, "inside bracket not allowed", true},
	{"noth,ing", docfmt.WordForbiddenRune, "non-trailing comma not allowed", true},
	{"som%thing", docfmt.WordForbiddenRune, "% in the middle of word not allowed", true},
	{"a%", docfmt.WordForbiddenRune, "trailing % not allowed", true},

	{exception, docfmt.ValidationOK, "excluded special word is ok", true},
	{exception + ",", docfmt.ValidationOK, "excluded special word is ok", true},
}

var exception = "SpeCiAL"

// add the word tests to the part tests.
func init() {
	for _, wt := range wordTests {
		if !wt.asPart {
			continue
		}

		var wantError []docfmt.ValidationResult
		if wt.want != docfmt.ValidationOK {
			wantError = []docfmt.ValidationResult{
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
			wantError []docfmt.ValidationResult
		}{
			name:      wt.comment,
			input:     wt.input,
			wantError: wantError,
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	for _, tt := range partTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotErr := docfmt.Validate(tt.input, exception)

			if !reflect.DeepEqual(gotErr, tt.wantError) {
				t.Errorf("Validate() error = %#v, want = %#v", gotErr, tt.wantError)
			}
		})
	}
}

func Test_validateWord(t *testing.T) {
	t.Parallel()

	for _, tt := range wordTests {
		t.Run(tt.comment, func(t *testing.T) {
			t.Parallel()

			if got := docfmt.ValidateWord(tt.input, map[string]struct{}{exception: {}}); got != tt.want {
				t.Errorf("ValidateWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
