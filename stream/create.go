package stream

import "os"

// FromNil creates a new IOStream that silences all output and provides no input.
func FromNil() IOStream {
	return NewIOStream(nil, nil, nil)
}

// FromEnv creates a new IOStream using the environment.
//
// The Stdin, Stdout and Stderr streams are used from the os package.
func FromEnv() IOStream {
	return IOStream{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}
