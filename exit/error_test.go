//spellchecker:words exit
package exit_test

//spellchecker:words errors testing pkglib exit testlib
import (
	"errors"
	"fmt"
	"testing"

	"go.tkw01536.de/pkglib/exit"
	"go.tkw01536.de/pkglib/testlib"
)

var (
	errStuff        = exit.NewErrorWithCode("stuff", 1)
	errStuffWrapped = fmt.Errorf("wrapping: %w", errStuff)
	errUnrelated    = errors.New("unrelated")
)

func TestCodeFromError(t *testing.T) {
	t.Parallel()

	var (
		errExitCode        = testlib.ProduceExitError(t, 1)
		errExitCodeWrapped = exit.FromExitError(errExitCode)
	)

	tests := []struct {
		name     string
		err      error
		generic  exit.ExitCode
		wantCode exit.ExitCode
		wantOK   bool
	}{
		{
			name:     "nil error returns zero value",
			err:      nil,
			generic:  10,
			wantCode: 0,
			wantOK:   true,
		},
		{
			name:     "Error object returns itself",
			err:      errStuff,
			generic:  10,
			wantCode: 1,
			wantOK:   true,
		},
		{
			name:     "Wrapped error returns same exit code",
			err:      errStuffWrapped,
			generic:  10,
			wantCode: 1,
			wantOK:   true,
		},
		{
			name:     "unrelated error returns invalid exit code",
			err:      errUnrelated,
			generic:  10,
			wantCode: 10,
			wantOK:   false,
		},
		{
			name:     "unwrapped exec error doesn't return exit code",
			err:      errExitCode,
			generic:  10,
			wantCode: 10,
			wantOK:   false,
		},
		{
			name:     "wrapped exec error returns exit code",
			err:      errExitCodeWrapped,
			generic:  10,
			wantCode: 1,
			wantOK:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotCode, gotOK := exit.CodeFromError(tt.err, tt.generic)
			if tt.wantCode != gotCode {
				t.Errorf("CodeFromError() code = %v, want %v", gotCode, tt.wantCode)
			}
			if tt.wantOK != gotOK {
				t.Errorf("CodeFromError() ok = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
