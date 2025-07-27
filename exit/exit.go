// Package exit defines exit and error behavior of programs and commands.
//
//spellchecker:words exit
package exit

//spellchecker:words math
import (
	"math"
	"os"
)

// ExitCode determines the exit behavior of a program.
// These are returned as an exit code to the operating system.
// See ExitCode.Return().
type ExitCode uint8

// Code turns an integer into an ExitCode safely.
// If it is outside of the range from the exit code, returns an unspecified non-zero value.
func Code(code int) ExitCode {
	if code < 0 || code > math.MaxUint8 {
		return math.MaxUint8
	}
	return ExitCode(code)
}

// Return returns this ExitCode to the operating system by invoking os.Exit().
func (code ExitCode) Return() {
	// NOTE: This function is untested
	os.Exit(int(code))
}

const (
	// ExitZero indicates that no error occurred.
	// It is the zero value of type ExitCode.
	ExitZero ExitCode = 0

	// ExitGeneric indicates a generic error occurred within this invocation.
	// This typically implies a subcommand-specific behavior wants to return failure to the caller.
	ExitGeneric ExitCode = 1

	// ExitUnknownCommand indicates that the user attempted to call a subcommand that is not defined.
	ExitUnknownCommand ExitCode = 2

	// ExitGeneralArguments indicates that the user attempted to pass invalid general arguments to the program.
	ExitGeneralArguments ExitCode = 3
	// ExitCommandArguments indicates that the user attempted to pass invalid command-specific arguments to a subcommand.
	ExitCommandArguments ExitCode = 4

	// ExitContext indicates an error with the underlying command context.
	ExitContext ExitCode = 254

	// ExitPanic indicates that the go code called panic() inside the execution of the current program.
	// This typically implies a bug inside a program.
	ExitPanic ExitCode = 255
)
