//go:build !nominify

//spellchecker:words minify
package minify

//spellchecker:words regexp github tdewolff minify html pkglib noop
import (
	"io"
	"regexp"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/svg"
	"github.com/tkw1536/pkglib/noop"
)

// minifier holds the minfier used for all html minification
//
// NOTE: We can't use an init function for this, because otherwise initialization order is incorrect.
var minifier = (func() *minify.M {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	return m
})()

// Minifiy returns a minifier that writes minification to dest.
// If minification is disabled, or no minifier for the given mediatype type exists, it returns a no-op wrapper around src.
//
// The caller must close the returned closer upon completion of writing.
func Minify(mediatype string, dest io.Writer) io.WriteCloser {
	_, _, f := minifier.Match(mediatype)
	if f == nil {
		return noop.Writer{Writer: dest}
	}
	return minifier.Writer(mediatype, dest)
}
