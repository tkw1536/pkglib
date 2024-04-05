package form_test

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/tkw1536/pkglib/httpx/form"
	"github.com/tkw1536/pkglib/httpx/form/field"
)

func ExampleForm_error() {

	// create a form that produces an error iff it is given non-nil values
	form := form.Form[string]{
		Fields: []field.Field{
			{Name: "error", Type: field.Text, Label: "Should an error be given?"},
		},

		Template: template.Must(template.New("form").Parse("{{ .ThisIsAnError }}")),

		Validate: func(r *http.Request, values map[string]string) (string, error) {
			if values["error"] != "" {
				return "", errors.New("debug error: error")
			}
			return "", nil
		},

		Success: func(data string, values map[string]string, w http.ResponseWriter, r *http.Request) error {
			return nil
		},
	}

	// call it
	fmt.Println("no error, no logger")
	makeFormRequest(&form, map[string]string{})

	fmt.Println("an error, no logger")
	makeFormRequest(&form, map[string]string{"error": "true"})

	// setup an error logger
	form.LogTemplateError = func(r *http.Request, err error) {
		fmt.Println("called error handler", err)
	}

	fmt.Println("no error, a logger")
	makeFormRequest(&form, map[string]string{})

	fmt.Println("an error, a logger")
	makeFormRequest(&form, map[string]string{"error": "true"})

	// call it again

	// Output: no error, no logger
	// an error, no logger
	// no error, a logger
	// an error, a logger
	// called error handler template: form:1:3: executing "form" at <.ThisIsAnError>: can't evaluate field ThisIsAnError in type form.FormContext
}
