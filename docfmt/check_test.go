//spellchecker:words docfmt
package docfmt

//spellchecker:words reflect testing github pkglib testlib
import (
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/testlib"
)

func TestAssertValid(t *testing.T) {
	t.Parallel()

	for _, tt := range partTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var wantPanic bool
			var wantError interface{}

			if Enabled {
				wantPanic = tt.wantError != nil
				if wantPanic {
					wantError = ValidationError{
						Message: tt.input,
						Results: tt.wantError,
					}
				}
			}

			gotPanic, gotError := testlib.DoesPanic(func() {
				AssertValid(tt.input)
			})

			if gotPanic != wantPanic {
				t.Errorf("AssertValid() got panic = %v, want = %v", gotPanic, wantPanic)
			}

			if !reflect.DeepEqual(gotError, wantError) {
				t.Errorf("AssertValid() got error = %v, want = %v", gotError, wantError)
			}
		})
	}
}
