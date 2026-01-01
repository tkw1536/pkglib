//spellchecker:words testlib
package testlib

//spellchecker:words errors exec runtime strconv testing
import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"go.tkw01536.de/pkglib/errorsx"
)

//spellchecker:words nosec subshell

const (
	minExitCode = 1   // minimum valid exit code
	maxExitCode = 127 // maximum valid exit code

	// total number of exit codes.
	exitCodeCount = maxExitCode - minExitCode + 1
)

// exitErrors holds a map that caches exit errors for each code.
// we intentionally hold a non-pointer, to force a copy.
var exitErrors map[int]func() exec.ExitError

func init() {
	// create a once value for each possible exit code.
	exitErrors = make(map[int]func() exec.ExitError, exitCodeCount)
	for i := minExitCode; i <= maxExitCode; i++ {
		exitErrors[i] = sync.OnceValue(func() exec.ExitError {
			return makeExitError(i)
		})
	}
}

// ProduceExitError returns an exec.ExitError that holds the given code.
//
// This function expects sh to be available in PATH, but also allows cmd on Windows.
// code must be in the range [1, 127] to be portable.
//
// Results from this function are cached to avoid having to execute a shell more than 127 times.
// If something goes wrong, this function panics.
func ProduceExitError(t *testing.T, code int) *exec.ExitError {
	t.Helper()

	if code < minExitCode || code > maxExitCode {
		panic(fmt.Sprintf("ProduceExitError: code must be in the range [1, 127]; got %d", code))
	}

	err := exitErrors[code]()
	return &err
}

// makeExitError makes a new exec.ExitError for the given code.
func makeExitError(code int) exec.ExitError {
	if code < minExitCode || code > maxExitCode {
		panic("never reached")
	}

	var cmd *exec.Cmd

	// look for sh first
	if _, err := exec.LookPath("sh"); err == nil {
		cmd = exec.CommandContext(context.Background(), "sh", "-c", "exit "+strconv.Itoa(code)) // #nosec: G204 inputs are guarded
	} else if runtime.GOOS == "windows" {
		// on windows only, fallback to "cmd"
		if _, err := exec.LookPath("cmd"); err == nil {
			cmd = exec.CommandContext(context.Background(), "cmd", "/C", "exit "+strconv.Itoa(code)) // #nosec: G204 inputs are guarded
		}
	}

	if cmd == nil {
		panic("makeExitError: neither 'sh' nor 'cmd' are available")
	}

	err := cmd.Run()

	exitErr, ok := errorsx.AsType[*exec.ExitError](err)
	if !ok {
		panic(fmt.Sprintf("makeExitError: produced type %T, expected *exec.ExitError", err))
	}
	got := exitErr.ExitCode()
	if got != code {
		panic(fmt.Sprintf("makeExitError: subshell invocation returned %d instead of %d", got, code))
	}
	return *exitErr
}
