// Package stream defines input and output streams.
package stream

import (
	"fmt"
	"io"

	"github.com/tkw1536/pkglib/nobufio"
)

// IOStream represents a set of input and output streams commonly associated to a process.
type IOStream struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}

// StdinIsATerminal checks if standard input is a terminal
func (str IOStream) StdinIsATerminal() bool {
	return nobufio.IsTerminal(str.Stdin)
}

// StdoutIsATerminal checks if standard output is a terminal
func (str IOStream) StdoutIsATerminal() bool {
	return nobufio.IsTerminal(str.Stdout)
}

// StderrIsATerminal checks if standard error is a terminal
func (str IOStream) StderrIsATerminal() bool {
	return nobufio.IsTerminal(str.Stderr)
}

// Printf is like [fmt.Printf] but prints to str.Stdout.
func (str IOStream) Printf(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(str.Stdout, format, args...)
}

// EPrintf is like [fmt.Printf] but prints to io.Stderr.
func (str IOStream) EPrintf(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(str.Stderr, format, args...)
}

// Print is like [fmt.Print] but prints to str.Stdout.
func (str IOStream) Print(args ...interface{}) (n int, err error) {
	return fmt.Fprint(str.Stdout, args...)
}

// EPrint is like [fmt.Print] but prints to str.Stderr.
func (str IOStream) EPrint(args ...interface{}) (n int, err error) {
	return fmt.Fprint(str.Stderr, args...)
}

// Println is like [fmt.Println] but prints to str.Stdout.
func (str IOStream) Println(args ...interface{}) (n int, err error) {
	return fmt.Fprintln(str.Stdout, args...)
}

// EPrintln is like [fmt.Println] but prints to io.Stderr.
func (str IOStream) EPrintln(args ...interface{}) (n int, err error) {
	return fmt.Fprintln(str.Stderr, args...)
}

// NewIOStream creates a new IOStream with the provided readers and writers.
// If any of them are set to nil, they are set to Null.
// When wrap is set to 0, it is set to a reasonable default.
//
// It furthermore wraps output as set by wrap.
func NewIOStream(Stdout, Stderr io.Writer, Stdin io.Reader) IOStream {
	if Stdout == nil {
		Stdout = Null
	}
	if Stderr == nil {
		Stderr = Null
	}
	if Stdin == nil {
		Stdin = Null
	}
	return IOStream{
		Stdin:  Stdin,
		Stdout: Stdout,
		Stderr: Stderr,
	}
}

// NonInteractive creates a new non-interactive writer from a single output stream.
//
// It is roughly equivalent to NewIOStream(Writer, Writer, nil, 0)
func NonInteractive(Writer io.Writer) IOStream {
	return NewIOStream(Writer, Writer, nil).NonInteractive()
}

// Streams creates a new IOStream with the provided streams and wrap.
// If any parameter is the zero value, copies the values from str.
func (str IOStream) Streams(Stdout, Stderr io.Writer, Stdin io.Reader, wrap int) IOStream {
	if Stdout == nil {
		Stdout = str.Stdout
	}
	if Stderr == nil {
		Stderr = str.Stderr
	}
	if Stdin == nil {
		Stdin = str.Stdin
	}
	return NewIOStream(Stdout, Stderr, Stdin)
}

// NonInteractive creates a new IOStream with [Null] as standard input.
func (str IOStream) NonInteractive() IOStream {
	return str.Streams(nil, nil, Null, 0)
}
