package httpx

import (
	"bytes"
	"net/http"
	"time"

	"github.com/tkw1536/pkglib/minify"
)

// Response represents a response to an http request.
type Response struct {
	ContentType string // defaults to text/plain
	Body        []byte

	Modtime    time.Time
	StatusCode int // defaults to a 2XX status code
}

// Minify returns a copy of the response with the content minified.
func (response Response) Minify() Response {
	response.Body = minify.MinifyBytes(response.ContentType, response.Body)
	return response
}

// Now returns a copy of the response with the current time set as the modtime.
func (response Response) Now() Response {
	response.Modtime = time.Now().UTC()
	return response
}

func (response Response) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if response.ContentType == "" {
		response.ContentType = "text/plain"
	}
	w.Header().Set("Content-Type", response.ContentType)

	// if we are responding with no status code, then we allow sending partial content
	// and also send proper headers.
	if response.StatusCode == 0 {
		http.ServeContent(w, r, "", response.Modtime, bytes.NewReader(response.Body))
		return
	}

	// else respond with a normal body only!
	if response.StatusCode < 0 {
		response.StatusCode = http.StatusOK
	}
	w.WriteHeader(response.StatusCode)
	w.Write(response.Body)
}
