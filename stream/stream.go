// Package stream defines input and output streams.
//
//spellchecker:words stream
package stream

//spellchecker:words github pkglib nobufio
import (
	"fmt"
	"io"

	"github.com/tkw1536/pkglib/nobufio"
)

//spellchecker:words nolint wrapcheck

// IOStream represents a set of input and output streams commonly associated to a process.
type IOStream struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}

// StdinIsATerminal checks if standard input is a terminal.
func (str IOStream) StdinIsATerminal() bool {
	return nobufio.IsTerminal(str.Stdin)
}

// StdoutIsATerminal checks if standard output is a terminal.
func (str IOStream) StdoutIsATerminal() bool {
	return nobufio.IsTerminal(str.Stdout)
}

// StderrIsATerminal checks if standard error is a terminal.
func (str IOStream) StderrIsATerminal() bool {
	return nobufio.IsTerminal(str.Stderr)
}

// Printf is like [fmt.Printf] but prints to str.Stdout.
func (str IOStream) Printf(format string, args ...any) (n int, err error) {
	return fmt.Fprintf(str.Stdout, format, args...) //nolint:wrapcheck  // don't wrap fmt here
}

// EPrintf is like [fmt.Printf] but prints to io.Stderr.
func (str IOStream) EPrintf(format string, args ...any) (n int, err error) {
	return fmt.Fprintf(str.Stderr, format, args...) //nolint:wrapcheck  // don't wrap fmt here
}

// Print is like [fmt.Print] but prints to str.Stdout.
func (str IOStream) Print(args ...any) (n int, err error) {
	return fmt.Fprint(str.Stdout, args...) //nolint:wrapcheck // don't wrap fmt here
}

// EPrint is like [fmt.Print] but prints to str.Stderr.
func (str IOStream) EPrint(args ...any) (n int, err error) {
	return fmt.Fprint(str.Stderr, args...) //nolint:wrapcheck  // don't wrap fmt here
}

// Println is like [fmt.Println] but prints to str.Stdout.
func (str IOStream) Println(args ...any) (n int, err error) {
	return fmt.Fprintln(str.Stdout, args...) //nolint:wrapcheck  // don't wrap fmt here
}

// EPrintln is like [fmt.Println] but prints to io.Stderr.
func (str IOStream) EPrintln(args ...any) (n int, err error) {
	return fmt.Fprintln(str.Stderr, args...) //nolint:wrapcheck  // don't wrap fmt here
}

// NewIOStream creates a new IOStream with the provided readers and writers.
// If any of them are set to nil, they are set to Null.
func NewIOStream(stdout, stderr io.Writer, stdin io.Reader) IOStream {
	if stdout == nil {
		stdout = Null
	}
	if stderr == nil {
		stderr = Null
	}
	if stdin == nil {
		stdin = Null
	}
	return IOStream{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
}

// It is roughly equivalent to NewIOStream(w, w, nil).
func NonInteractive(w io.Writer) IOStream {
	return NewIOStream(w, w, nil).NonInteractive()
}

// Streams creates a new IOStream with the provided streams.
// If any parameter is the zero value, copies the values from str.
func (str IOStream) Streams(stdout, stderr io.Writer, stdin io.Reader, wrap int) IOStream {
	if stdout == nil {
		stdout = str.Stdout
	}
	if stderr == nil {
		stderr = str.Stderr
	}
	if stdin == nil {
		stdin = str.Stdin
	}
	return NewIOStream(stdout, stderr, stdin)
}

// NonInteractive creates a new IOStream with [Null] as standard input.
func (str IOStream) NonInteractive() IOStream {
	return str.Streams(nil, nil, Null, 0)
}
