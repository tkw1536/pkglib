//go:build !nominify

//spellchecker:words minify
package minify

//spellchecker:words sync github tdewolff minify html json pkglib noop
import (
	"io"
	"sync"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	"github.com/tkw1536/pkglib/noop"
)

//spellchecker:words minifier nominify

// minifier holds the minifier used for all html minification
//
// NOTE: We can't use an init function for this, because otherwise initialization order is incorrect.
var minifier = sync.OnceValue(func() *minify.M {
	m := minify.New()

	for _, typ := range []string{
		"text/html",
		"application/html",
		"application/xhtml+xml",
	} {
		m.AddFunc(typ, html.Minify)
	}

	for _, typ := range []string{
		"text/css",
		"application/css",
	} {
		m.AddFunc(typ, css.Minify)
	}

	for _, typ := range []string{
		"image/svg+xml",
		"application/svg+xml",
		"text/svg+xml",
	} {
		m.AddFunc(typ, svg.Minify)
	}

	for _, typ := range []string{
		"application/json",
		"text/json",
	} {
		m.AddFunc(typ, json.Minify)
	}

	for _, typ := range []string{
		"application/javascript",
		"application/ecmascript",
		"application/x-javascript",
		"application/x-ecmascript",
		"text/javascript",
		"text/ecmascript",
		"text/x-javascript",
		"text/x-ecmascript",
	} {
		m.AddFunc(typ, js.Minify)
	}

	for _, typ := range []string{
		"application/xml",
		"text/xml",
		"application/x-xml",
		"application/xml;charset=utf-8",
		"text/xml;charset=utf-8",
		"application/xml;charset=utf-16",
		"text/xml;charset=utf-16",
	} {
		m.AddFunc(typ, xml.Minify)
	}

	return m
})

// Minify returns a minifier that writes minification to dest.
// If minification is disabled, or no minifier for the given media type exists, it returns a no-op wrapper around src.
//
// The caller must close the returned closer upon completion of writing.
func Minify(mediaType string, dest io.Writer) io.WriteCloser {
	m := minifier()
	_, _, f := m.Match(mediaType)
	if f == nil {
		return noop.Writer{Writer: dest}
	}
	return m.Writer(mediaType, dest)
}
