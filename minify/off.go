//go:build nominify

package minify

import "io"

// Minifiy returns a minifier that writes minification to dest.
// If minification is disabled, or no minifier for the given mediatype exists, it returns a no-op wrapper around src.
//
// The caller must close the returned closer upon completion of writing.
func Minify(mediatype string, dest io.Writer) io.WriteCloser {
	return noop{Writer: dest}
}
