//spellchecker:words exit
package exit

//spellchecker:words pkglib stream
import (
	"fmt"

	"go.tkw01536.de/pkglib/stream"
)

var errUnknown = NewErrorWithCode("unknown error", ExitGeneric)

// Die prints a non-nil err to io.Stderr and returns an error with an exit code.
// If err is nil, it does nothing and returns nil.
func Die(str stream.IOStream, err error) error {
	// fast case: not an error
	if err == nil {
		return nil
	}

	// if we do not have a code, wrap the error in it!
	if _, ok := CodeFromError(err); !ok {
		err = fmt.Errorf("%w: %w", errUnknown, err)
	}

	// print the error message to standard error in a wrapped way
	if message := fmt.Sprint(err); message != "" {
		_, _ = str.EPrintln(message) // no way to report the failure
	}

	return err
}
