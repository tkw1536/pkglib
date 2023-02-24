package minify

import (
	"bytes"
	"io"
)

type noop struct {
	io.Writer
}

func (noop) Close() error {
	return nil
}

func MinifyBytes(mediatype string, in []byte) []byte {
	// create a new minifier
	var buffer bytes.Buffer
	writer := Minify(mediatype, &buffer)

	// write and then close it!
	if _, err := writer.Write(in); err != nil {
		return in
	}

	if err := writer.Close(); err != nil {
		return in
	}

	// return the bytes!
	return buffer.Bytes()

}
