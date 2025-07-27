//spellchecker:words exit
package exit_test

//spellchecker:words errors testing pkglib exit
import (
	"errors"
	"fmt"
	"testing"

	"go.tkw01536.de/pkglib/exit"
)

var (
	errStuff        = exit.NewErrorWithCode("stuff", exit.ExitGeneric)
	errStuffWrapped = fmt.Errorf("wrapping: %w", errStuff)
	errUnrelated    = errors.New("unrelated")
)

func TestCodeFromError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode exit.ExitCode
		wantOK   bool
	}{
		{
			name:     "nil error returns zero value",
			err:      nil,
			wantCode: exit.ExitZero,
			wantOK:   true,
		},
		{
			name:     "Error object returns itself",
			err:      errStuff,
			wantCode: exit.ExitGeneric,
			wantOK:   true,
		},
		{
			name:     "Wrapped error returns same exit code",
			err:      errStuffWrapped,
			wantCode: exit.ExitGeneric,
			wantOK:   true,
		},
		{
			name:     "unrelated error returns invalid exit code",
			err:      errUnrelated,
			wantCode: exit.ExitGeneric,
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotCode, gotOK := exit.CodeFromError(tt.err)
			if tt.wantCode != gotCode {
				t.Errorf("CodeFromError() code = %v, want %v", gotCode, tt.wantCode)
			}
			if tt.wantOK != gotOK {
				t.Errorf("CodeFromError() ok = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
