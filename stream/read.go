//spellchecker:words stream
package stream

//spellchecker:words github pkglib nobufio
import "github.com/tkw1536/pkglib/nobufio"

//spellchecker:words nolint wrapcheck

// ReadLine is like [nobufio.ReadLine] on the standard input.
func (str IOStream) ReadLine() (string, error) {
	return nobufio.ReadLine(str.Stdin) //nolint:wrapcheck // don't wrap nobufio errors
}

// ReadPassword is like [nobufio.ReadPassword] on the standard input.
func (str IOStream) ReadPassword() (string, error) {
	return nobufio.ReadPassword(str.Stdin) //nolint:wrapcheck // don't wrap nobufio errors
}

// ReadPasswordStrict is like [nobufio.ReadPasswordStrict] on the standard input.
func (str IOStream) ReadPasswordStrict() (string, error) {
	return nobufio.ReadPasswordStrict(str.Stdin) //nolint:wrapcheck // don't wrap nobufio errors
}
