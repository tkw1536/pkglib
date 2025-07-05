//spellchecker:words minify
package minify

//spellchecker:words bytes github pkglib errorsx
import (
	"bytes"
	"fmt"

	"go.tkw01536.de/pkglib/errorsx"
)

// MinifyBytes minifies the bytes described.
// If an error occurs, returns the input unchanged.
func MinifyBytes(mediaType string, in []byte) []byte {
	var buffer bytes.Buffer

	if err := minifyInto(mediaType, &buffer, in); err != nil {
		return in
	}
	return buffer.Bytes()
}

func minifyInto(mediaType string, buf *bytes.Buffer, in []byte) (e error) {
	writer := Minify(mediaType, buf)
	defer errorsx.Close(writer, &e, "writer")

	if _, err := writer.Write(in); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}
