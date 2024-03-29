// Package form provides a form abstraction for http
package form

import (
	"html/template"
	"net/http"
	"strings"

	_ "embed"

	"github.com/gorilla/csrf"
	"github.com/tkw1536/pkglib/httpx"
	"github.com/tkw1536/pkglib/httpx/content"
	"github.com/tkw1536/pkglib/httpx/form/field"
)

// Form provides a form that a user can submit via a http POST method call.
// It implements [http.Handler], see [Form.ServeHTTP] for details on how form submission works.
//
// Data is the type of data the form parses.
type Form[Data any] struct {
	// Fields are the set of fields to be included in this form.
	Fields []field.Field

	// FieldTemplate is an optional template to be executed for each field.
	// FieldTemplate may be nil; in which case [DefaultFieldTemplate] is used.
	FieldTemplate *template.Template

	// SkipCSRF if CSRF should be explicitly omitted
	SkipCSRF bool

	// Skip determines if the form can be skipped and a result page can be rendered directly.
	// If so, values for the result have to be passed out of the request.
	//
	// A nil Skip is assumed to not allow the form to be skipped.
	Skip func(*http.Request) (data Data, skip bool)

	// Template represents the template to render for GET requests.
	// It is passed the return value of [TemplateContext].
	Template *template.Template

	// TemplateContext is the context to be used for Template.
	// A nil TemplateContext function returns the FormContext object as-is.
	TemplateContext func(FormContext, *http.Request) any

	// Validate is a function that validates submitted form values, and parses them into a data object.
	// See [Form.Values] on how this is used.
	//
	// A nil validation function is assumed to return the zero value of Data and nil error.
	Validate func(r *http.Request, values map[string]string) (Data, error)

	// Success is a function that renders a successfully parsed form (either via [Validate] or [SkipForm]) into a response.
	// A nil [Success] function is assumed to just render the [Template] above.
	Success func(data Data, values map[string]string, w http.ResponseWriter, r *http.Request) error
}

// HTML renders the values for the given html fields into a html template.
// IsError indicates if fields with the EmptyOnError flag should be omitted.
func (form *Form[D]) HTML(values map[string]string, IsError bool) template.HTML {
	var builder strings.Builder

	for _, field := range form.Fields {
		value := values[field.Name]
		if IsError && field.EmptyOnError {
			value = ""
		}

		field.WriteTo(&builder, form.FieldTemplate, value)
	}

	return template.HTML(builder.String())
}

// Values validates values inside the given request, and returns parsed out form values from a post request.
//
// Validation is performed using the [Form.Validate] function.
// Validation is passed the extracted field values.
//
// - Only values corresponding to a form field in the request are used.
// - If multiple values are submitted for a specific field, only the first one is included.
// - If a value is missing, it is assigned the empty string.
//
// If the parsed out values do not match, an error is returned instead.
//
// Upon return, the map holds parsed out field values,
// Err indicates if an error occurred.
func (form *Form[Data]) Values(r *http.Request) (map[string]string, Data, error) {
	// parse the form
	if err := r.ParseForm(); err != nil {
		var data Data
		return nil, data, err
	}

	// pick each of the values
	values := make(map[string]string, len(form.Fields))
	for _, field := range form.Fields {
		values[field.Name] = r.PostForm.Get(field.Name)
	}

	// validate the form (if any)
	var data Data
	var err error
	if form.Validate != nil {
		data, err = form.Validate(r, values)
		if err != nil {
			return values, data, err
		}
	}

	// return the values
	return values, data, nil
}

// ServeHTTP implements serving the http form.
//
// This works in two stages.
//
// In the first stage, values for the form are parsed.
//
// - If the request method is post, the form values are extracted using the [Values] method.
// - If the request method is get, and [Form.Skip] is defined, check if the form can be skipped.
// - If the request method is get, and the form cannot be skipped, we continue to the second stage below.
// - For all other cases, [ErrMethodNotAllowed] is rendered.
//
// In the second stage, we either render the form template, or the success template.
//
// - If form data was generated successfully (either via [Values] or [SkipForm]), we invoke [Form.RenderSuccess] with the appropriate data.
// - Otherwise, we render the Form Template with an appropriate error.
func (form *Form[Data]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Allow interception of form stuff
	switch {
	default:
		httpx.ErrMethodNotAllowed.ServeHTTP(w, r)
		return
	case r.Method == http.MethodPost:
		values, data, err := form.Values(r)
		if err != nil {
			form.renderForm(err, values, w, r)
		} else {
			form.renderSuccess(data, values, w, r)
		}
	case r.Method == http.MethodGet && form.Skip != nil:
		if data, skip := form.Skip(r); skip {
			form.renderSuccess(data, nil, w, r)
			return
		}
		fallthrough
	case r.Method == http.MethodGet:
		form.renderForm(nil, nil, w, r)
	}
}

func (form *Form[Data]) renderForm(err error, values map[string]string, w http.ResponseWriter, r *http.Request) {
	template := form.HTML(values, err != nil)
	if !form.SkipCSRF {
		template += csrf.TemplateField(r)
	}

	ctx := FormContext{Err: err, Form: template}

	// must have a form or a RenderForm
	if form.Template == nil {
		panic("form.Template is nil")
	}

	// get the template context
	var tplctx any
	if form.TemplateContext == nil {
		tplctx = ctx
	} else {
		tplctx = form.TemplateContext(ctx, r)
	}

	// render the form
	content.WriteHTML(tplctx, nil, form.Template, w, r)
}

// FormContext is passed to [Form.TemplateContext] when used
type FormContext struct {
	// Error is the underlying error (if any)
	Err error

	// Template is the underlying template rendered as html
	Form template.HTML
}

// Error returns the underlying error string.
// If Err is nil, it returns an empty string.
func (fc FormContext) Error() string {
	if fc.Err == nil {
		return ""
	}
	return fc.Err.Error()
}

// renderSuccess renders a successful pass of the form
// if an error occurs during rendering, renderForm is called instead
func (form *Form[D]) renderSuccess(data D, values map[string]string, w http.ResponseWriter, r *http.Request) {
	err := form.Success(data, values, w, r)
	if err == nil {
		return
	}
	form.renderForm(err, values, w, r)
}

//go:embed "form.html"
var formBytes []byte

// FormTemplate is a template to embed a form
var FormTemplate = template.Must(template.New("form.html").Parse(string(formBytes)))
