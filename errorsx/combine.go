//spellchecker:words errorsx
package errorsx

//spellchecker:words errors
import (
	"errors"
)

//spellchecker:words unsafe

// Combine is like [errors.Join] except if there is only a single non-nil error, it is returned unchanged.
func Combine(errs ...error) error {
	var anError error
	for _, err := range errs {
		if err == nil {
			continue
		}

		// saw a second error
		if anError != nil {
			return errors.Join(errs...)
		}

		anError = err
	}

	// we saw at most one error => return it
	return anError
}
