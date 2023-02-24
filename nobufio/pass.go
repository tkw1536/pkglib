package nobufio

import (
	"errors"
	"io"

	"golang.org/x/term"
)

// ReadPassword is like ReadLine, except that it turns off terminal echo.
// When reader is not a terminal, behaves like [ReadLine]
func ReadPassword(reader io.Reader) (value string, err error) {
	value, err = ReadPasswordStrict(reader)
	if err == ErrNoTerminal {
		return ReadLine(reader)
	}
	return
}

// ErrNoTerminal is returned by ReadPasswordStrict when stdin is not a terminal
var ErrNoTerminal = errors.New("reader is not a terminal")

// ReadPasswordSrict is like ReadPassword, except that when reader is not a terminal, returns ErrNoTerminal.
func ReadPasswordStrict(reader io.Reader) (value string, err error) {
	// check if reader is a terminal
	file, ok := reader.(interface{ Fd() uintptr })
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return "", ErrNoTerminal
	}

	// read the bytes
	bytes, err := term.ReadPassword(int(file.Fd()))
	return string(bytes), err
}
