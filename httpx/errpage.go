package httpx

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime/debug"
	"strings"

	_ "embed"
)

// spellchecker:words errpage

// RenderErrorPage renders a debug error page instead of the fallback response res.
// The error page is intended to replace error pages for debugging and should not be used in production.
//
// It will keep the original status code of res, but will replace the content type and body.
// It will render as 'text/html'.
func RenderErrorPage(err error, res Response, w http.ResponseWriter, r *http.Request) {
	newErrorPage(err, r).Response(res).Minify().ServeHTTP(w, r)
}

// newErrorPage returns a new error page for the specified error and request.
func newErrorPage(err error, r *http.Request) (page errorPage) {
	page.Stack = string(debug.Stack())

	page.Error = newSError(err)

	if r != nil {
		page.Method = r.Method
		if r.URL != nil {
			page.Path = r.URL.Path
		}
		page.Headers = r.Header
	}

	return
}

// errorPage represents a debug error page to be shown to the user.
// For formatting of the error page, see [FormatHTML].
type errorPage struct {
	Method  string
	Path    string
	Headers http.Header

	// Stack is the stack as returned by [runtime/debug.Stack]
	Stack string

	// Error is the underlying error cause
	Error sError
}

// Response replaces the body and content type of the given response by the formatted html.
func (err errorPage) Response(res Response) Response {
	res.Body = []byte(err.FormatHTML(res))
	res.ContentType = ContentTypeHTML
	return res
}

//go:embed errpage.html
var errpageHTML string
var errpageHTMLTemplate = template.Must(template.New("errpage.html").Parse(errpageHTML))

// context passed to error page template
type errPageContext struct {
	Error    errorPage
	Original Response
}

func (epc errPageContext) BodyString() string {
	return string(epc.Original.Body)
}

// FormatHTML formats the error page as html.
func (err errorPage) FormatHTML(res Response) template.HTML {
	var builder strings.Builder

	// TODO: Ignores template errors
	_ = errpageHTMLTemplate.Execute(&builder, errPageContext{Error: err, Original: res})

	return template.HTML(builder.String())
}

// sError represents a stringified error
type sError struct {
	Error  string // Error is the result of calling the Error() method
	Type   string // Type is the type of error
	Source string // Source is the result of formatting the error as a go source

	Unwrap []sError // The source of wrapped errors
}

// newSError safely turns an error into an error
func newSError(err error) sError {
	e := sError{
		Error:  fmt.Sprintf("%s", err),
		Type:   fmt.Sprintf("%T", err),
		Source: fmt.Sprintf("%#v", err),
	}

	// find the child errors
	var children []error
	switch x := err.(type) {
	case interface{ Unwrap() error }:
		children = []error{x.Unwrap()}
	case interface{ Unwrap() []error }:
		children = x.Unwrap()
	}

	// turn non-nil errors into NewError objects
	e.Unwrap = make([]sError, 0, len(children))
	for _, err := range children {
		if err == nil {
			continue
		}
		e.Unwrap = append(e.Unwrap, newSError(err))
	}

	// and we've built the error
	return e
}
