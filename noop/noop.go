// noop provides reader and writer
//
//spellchecker:words noop
package noop

import "io"

// Writer extends an io.Writer by adding a noop close operation
type Writer struct {
	io.Writer
}

func (Writer) Close() error {
	return nil
}

// Reader extends an io.Reader by adding a noop close operation
type Reader struct {
	io.Reader
}

func (Reader) Close() error {
	return nil
}
