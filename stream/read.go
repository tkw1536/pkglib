package stream

import "github.com/tkw1536/pkglib/nobufio"

// ReadLine is like [nobufio.ReadLine] on the standard input
func (str IOStream) ReadLine() (string, error) {
	return nobufio.ReadLine(str.Stdin)
}

// ReadPassword is like [nobufio.ReadPassword] on the standard input
func (str IOStream) ReadPassword() (string, error) {
	return nobufio.ReadPassword(str.Stdin)
}

// ReadPasswordStrict is like [nobufio.ReadPasswordStrict] on the standard input
func (str IOStream) ReadPasswordStrict() (string, error) {
	return nobufio.ReadPasswordStrict(str.Stdin)
}
