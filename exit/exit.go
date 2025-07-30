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
// See [ExitCode.Return].
type ExitCode uint8

// Code turns an integer into an [ExitCode] safely.
// If it is outside of the range from the exit code, returns an unspecified non-zero value.
func Code(code int) ExitCode {
	if code < 0 || code > math.MaxUint8 {
		return math.MaxUint8
	}
	return ExitCode(code)
}

// Return returns this ExitCode to the operating system by invoking [os.Exit].
func (code ExitCode) Return() {
	// NOTE: This function is untested
	os.Exit(int(code))
}
