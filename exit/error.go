//spellchecker:words exit
package exit

//spellchecker:words exec pkglib errorsx
import (
	"os/exec"

	"go.tkw01536.de/pkglib/errorsx"
)

// errorWithCode is an error that holds an exit code.
type errorWithCode interface {
	error
	exitCode() ExitCode
}

var (
	_ errorWithCode = &codeError{}
	_ errorWithCode = &exitError{}
)

// CodeFromError returns the ExitCode contained in error, if any.
// See [NewErrorWithCode] and [ErrorFromExit].
// The exit code is found by [errors.As] unwrapping into an error created by this package.
//
// When err is nil, returns code 0.
// When err does not hold any exit code, returns the provided generic code and false.
func CodeFromError(err error, generic ExitCode) (code ExitCode, ok bool) {
	if err == nil {
		return 0, true
	}
	if codeErr, ok := errorsx.AsType[errorWithCode](err); ok {
		return codeErr.exitCode(), true
	}
	return generic, false
}

// NewErrorWithCode creates a new error that additionally holds the given exit code.
func NewErrorWithCode(message string, code ExitCode) error {
	return &codeError{message: message, code: code}
}

type codeError struct {
	code    ExitCode
	message string
}

func (err *codeError) exitCode() ExitCode {
	return err.code
}

func (err *codeError) Error() string {
	return err.message
}

// FromExitError create a new error wrapping an [exec.ExitError].
// The private interface is guaranteed to be implemented by [exec.ExitError].
// The returned error holds the appropriate exit code.
func FromExitError(err *exec.ExitError) error {
	if err == nil {
		return nil
	}
	return &exitError{err: err}
}

type exitError struct {
	err *exec.ExitError
}

func (err *exitError) Error() string {
	return err.err.Error()
}

func (err *exitError) Unwrap() error {
	return err.err
}

func (err *exitError) exitCode() ExitCode {
	return Code(err.err.ExitCode())
}
