package docfmt

import (
	"fmt"
	"strings"
)

// AssertValid asserts that message is properly formatted and calling Validate on it returns no results.
//
// When checking is disabled, no runtime checking is performed.
// When checking is enabled and a message fails to pass validation, calls panic()
func AssertValid(message string) {
	if Enabled {
		if errors := Validate(message); len(errors) != 0 {
			panic(ValidationError{
				Message: message,
				Results: errors,
			})
		}
	}
}

// ValidationError is returned when a message fails validation.
// It implements the built-in error interface.
type ValidationError struct {
	Results []ValidationResult

	// message is the message being checked
	Message string
}

func (ve ValidationError) Error() string {
	// NOTE(twiesing): This function is untested because it is used only for developing

	messages := make([]string, len(ve.Results))
	for i, res := range ve.Results {
		messages[i] = res.Error()
	}

	return fmt.Sprintf("message %q failed validation: %s", ve.Message, strings.Join(messages, "\n"))
}

// AssertValidArgs checks that args does not contain exactly one argument of type error.
//
// When checking is disabled, no runtime checking is performed.
// When checking is enabled and the check is failed, calls panic()
func AssertValidArgs(args ...any) {
	if Enabled {
		if len(args) != 1 {
			return
		}
		if _, ok := args[0].(error); ok {
			panic("AssertValidArgs: single error argument provided")
		}

	}
}
