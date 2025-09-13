// Package exit defines exit and error behavior of programs and commands.
//
//spellchecker:words exit
package exit_test

//spellchecker:words errors math exec strconv testing pkglib exit
import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"go.tkw01536.de/pkglib/exit"
)

//spellchecker:words nosec outbounds

func TestCode_inbounds(t *testing.T) {
	t.Parallel()

	for i := range math.MaxUint8 {
		want := exit.ExitCode(i) // #nosec G115 // in bounds by loop condition
		if got := exit.Code(i); got != want {
			t.Errorf("Code(%d) = %v, want %v", i, got, want)
		}
	}
}

func TestCode_outbounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
		want exit.ExitCode
	}{
		{"negative number", -1, math.MaxUint8},
		{"too big positive number", math.MaxUint8 + 1, math.MaxUint8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := exit.Code(tt.code); got != tt.want {
				t.Errorf("Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

// environment variable used to setup a pkglib test that exits with a specific code.
const exitCodeEnv = "PKGLIB_TEST_EXIT_CODE"

func TestExitCode_Return(t *testing.T) {
	t.Parallel()

	if exitCodeStr := os.Getenv(exitCodeEnv); exitCodeStr != "" {
		var code exit.ExitCode
		if _, err := fmt.Sscanf(exitCodeStr, "%d", &code); err != nil {
			t.Fatalf("Failed to parse exit code: %v", err)
		}
		code.Return()
		return
	}

	for exitCode := range uint8(255) {
		exitCodeStr := strconv.FormatUint(uint64(exitCode), 10)
		t.Run(exitCodeStr, func(t *testing.T) {
			t.Parallel()

			// invoke the current test executable with the exit code
			cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run="+t.Name()) // #nosec G204 -- we need this for the test
			cmd.Env = append(os.Environ(), exitCodeEnv+"="+exitCodeStr)

			var gotCode int
			var exitErr *exec.ExitError
			if errors.As(cmd.Run(), &exitErr) {
				gotCode = exitErr.ExitCode()
			}
			if gotCode != int(exitCode) {
				t.Errorf("ExitCode.Return() exited with code %d, want %d", gotCode, exitCode)
			}
		})
	}
}
