//go:build nominify

//spellchecker:words minify
package minify

//spellchecker:words github pkglib noop
import (
	"io"

	"go.tkw01536.de/pkglib/noop"
)

// spellchecker:words minifier nominify

// Minify returns a minifier that writes minification to dest.
// If minification is disabled, or no minifier for the given mediaType exists, it returns a no-op wrapper around src.
//
// The caller must close the returned closer upon completion of writing.
func Minify(mediaType string, dest io.Writer) io.WriteCloser {
	return noop.Writer{Writer: dest}
}
