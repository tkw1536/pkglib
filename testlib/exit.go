//spellchecker:words testlib
package testlib

//spellchecker:words errors exec runtime strconv
import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// ProduceExitError returns an exec.ExitError that holds the given code.
// This function expects sh to be available in PATH, but also allows cmd on Windows.
// code must be in the range [1, 127] to be portable.
//
// If something goes wrong, this function panics.
func ProduceExitError(code int) *exec.ExitError {
	if code < 1 || code > 127 {
		panic(fmt.Sprintf("ProduceExitError: code must be in the range [1, 127]; got %d", code))
	}

	var cmd *exec.Cmd

	// look for sh first
	if _, err := exec.LookPath("sh"); err == nil {
		cmd = exec.Command("sh", "-c", "exit "+strconv.Itoa(code)) // #nosec: G204 inputs are guarded
	} else if runtime.GOOS == "windows" {
		// on windows only, fallback to "cmd"
		if _, err := exec.LookPath("cmd"); err == nil {
			cmd = exec.Command("cmd", "/C", "exit "+strconv.Itoa(code)) // #nosec: G204 inputs are guarded
		}
	}

	if cmd == nil {
		panic("ProduceExitError: neither 'sh' nor 'cmd' are available")
	}

	err := cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		panic(fmt.Sprintf("ProduceExitError: produced type %T, expected *exec.ExitError", err))
	}
	got := exitErr.ExitCode()
	if got != code {
		panic(fmt.Sprintf("ProduceExitError: subshell invocation returned %d instead of %d", got, code))
	}
	return exitErr
}
