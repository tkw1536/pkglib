// Package httpx provides additional [http.Handler]s and utility functions
package httpx

import (
	"bytes"
	"net/http"
	"time"

	"github.com/tkw1536/pkglib/minify"
)

// Response represents a static http Response.
// It implements [http.Handler].
type Response struct {
	ContentType string // defaults to text/plain
	Body        []byte

	Modtime    time.Time
	StatusCode int // defaults to a 2XX status code
}

// Minify returns a copy of the response with minified content.
func (response Response) Minify() Response {
	response.Body = minify.MinifyBytes(response.ContentType, response.Body)
	return response
}

// Now returns a copy of the response with the Modtime field set to the current time in UTC.
func (response Response) Now() Response {
	response.Modtime = time.Now().UTC()
	return response
}

func (response Response) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// setup and send the ContentType header iff it is set
	if response.ContentType == "" {
		response.ContentType = "text/plain"
	}
	w.Header().Set("Content-Type", response.ContentType)

	// when no status code is set use [http.ServeContent]
	// which is way better than anything we could implement
	if response.StatusCode == 0 {
		http.ServeContent(w, r, "", response.Modtime, bytes.NewReader(response.Body))
		return
	}

	// ensure that StatusCode is valid!
	if response.StatusCode < 0 {
		response.StatusCode = http.StatusOK
	}

	// write only the response with the given content type
	w.WriteHeader(response.StatusCode)
	w.Write(response.Body)
}
