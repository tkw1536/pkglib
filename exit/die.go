//spellchecker:words exit
package exit

import (
	"fmt"
	"io"
)

// Die prints a non-nil err to w and returns an error with an exit code.
// An error without an error code is wrapped with wrap, which should hold an exit code.
// If err is nil, it does nothing and returns nil.
func Die(w io.Writer, err error, wrap error) error {
	// fast case: not an error
	if err == nil {
		return nil
	}

	// if we do not have a code, wrap the error.
	// The generic exit code passed here is discarded.
	if _, ok := CodeFromError(err, 1); !ok {
		err = fmt.Errorf("%w: %w", wrap, err)
	}

	// print the error message to standard error in a wrapped way
	if message := fmt.Sprint(err); message != "" {
		_, _ = fmt.Fprintln(w, message) // no way to report the failure
	}

	return err
}
