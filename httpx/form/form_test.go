//spellchecker:words form
package form_test

//spellchecker:words errors html template http strconv testing github pkglib httpx form field
import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"testing"

	"go.tkw01536.de/pkglib/httpx/form"
	"go.tkw01536.de/pkglib/httpx/form/field"
)

func TestForm_logger(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		ShouldValidate bool
		ShouldSuccess  bool
		WantCalled     bool
	}{
		{ShouldValidate: true, ShouldSuccess: true, WantCalled: false},
		{ShouldValidate: true, ShouldSuccess: false, WantCalled: true},
		{ShouldValidate: false, ShouldSuccess: true, WantCalled: true},
		{ShouldValidate: false, ShouldSuccess: false, WantCalled: true},
	} {
		t.Run(fmt.Sprintf("Validate %t Success %t", tt.ShouldValidate, tt.ShouldSuccess), func(t *testing.T) {
			t.Parallel()

			frm := makeTestForm(t)
			frm.Template = template.Must(template.New("form").Parse("{{ .ThisIsAnError }}"))

			// setup a LogTemplateError that records if it was called or not
			called := false
			frm.LogTemplateError = func(r *http.Request, err error) {
				called = true
			}

			makeFormRequest(t, &frm, map[string]string{"validate": strconv.FormatBool(tt.ShouldValidate), "success": strconv.FormatBool(tt.ShouldSuccess)})

			if called != tt.WantCalled {
				t.Errorf("want called = %t, got called = %t", tt.WantCalled, called)
			}
		})
	}
}

func TestForm_formContext_afterSuccess(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		ShouldValidate bool
		ShouldSuccess  bool

		WantCalled       bool
		WantAfterSuccess bool
	}{
		{ShouldValidate: true, ShouldSuccess: true, WantCalled: false},
		{ShouldValidate: true, ShouldSuccess: false, WantCalled: true, WantAfterSuccess: true},
		{ShouldValidate: false, ShouldSuccess: true, WantCalled: true, WantAfterSuccess: false},
		{ShouldValidate: false, ShouldSuccess: false, WantCalled: true, WantAfterSuccess: false},
	} {
		t.Run(fmt.Sprintf("Validate %t Success %t", tt.ShouldValidate, tt.ShouldSuccess), func(t *testing.T) {
			t.Parallel()

			frm := makeTestForm(t)
			frm.Template = template.Must(template.New("form").Parse("{{ .Form }}"))

			// record if the TemplateContext function is called
			// and record what AfterSuccess was like
			called := false
			afterSuccess := false
			frm.TemplateContext = func(fc form.FormContext, r *http.Request) any {
				called = true
				afterSuccess = fc.AfterSuccess
				return fc
			}

			makeFormRequest(t, &frm, map[string]string{"validate": strconv.FormatBool(tt.ShouldValidate), "success": strconv.FormatBool(tt.ShouldSuccess)})

			if called != tt.WantCalled {
				t.Errorf("want called = %t, got called = %t", tt.WantCalled, called)
			}
			if called && afterSuccess != tt.WantAfterSuccess {
				t.Errorf("want afterSuccess = %t, got afterSuccess = %t", tt.WantAfterSuccess, afterSuccess)
			}
		})
	}
}

var (
	errSuccess  = errors.New("<success>")
	errValidate = errors.New("<validate>")
)

// testForm makes a form that can pass or fail the validate and success stages.
func makeTestForm(t *testing.T) form.Form[bool] {
	t.Helper()

	return form.Form[bool]{
		Fields: []field.Field{
			{Name: "validate", Type: field.Text, Label: "Should the validate stage be passed?"},
			{Name: "success", Type: field.Text, Label: "Should the success stage be passed?"},
		},

		Validate: func(r *http.Request, values map[string]string) (bool, error) {
			if validate := values["validate"]; validate == "" || validate == "false" {
				return false, errValidate
			}

			success := values["success"]
			return success != "" && success != "false", nil
		},

		Success: func(data bool, values map[string]string, w http.ResponseWriter, r *http.Request) error {
			if !data {
				return errSuccess
			}
			return nil
		},
	}
}
