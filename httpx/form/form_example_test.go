//spellchecker:words form
package form_test

//spellchecker:words errors html template http httptest strings testing github pkglib httpx content form field
import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/tkw1536/pkglib/httpx/content"
	"github.com/tkw1536/pkglib/httpx/form"
	"github.com/tkw1536/pkglib/httpx/form/field"
)

func ExampleForm() {
	formTemplate := template.Must(template.New("form").Parse("<!doctype html><title>Form</title>{{ if .Error }}<p>Error: {{ .Error }}</p>{{ end }}{{ .Form }}"))
	successTemplate := template.Must(template.New("success").Parse("<!doctype html><title>Success</title>Welcome {{ . }}"))

	form := form.Form[string]{
		Fields: []field.Field{
			{Name: "givenName", Type: field.Text, Label: "Given Name"},
			{Name: "familyName", Type: field.Text, Label: "Family Name"},
			{Name: "password", Type: field.Password, Label: "Password", EmptyOnError: true},
		},

		Template:        formTemplate,
		TemplateContext: func(fc form.FormContext, r *http.Request) any { return fc },

		Validate: func(r *http.Request, values map[string]string) (string, error) {
			given, family, password := values["givenName"], values["familyName"], values["password"]
			if given == "" {
				return "", errors.New("given name must not be empty")
			}
			if family == "" {
				return "", errors.New("family name must not be empty")
			}
			if password == "" {
				return "", errors.New("no password provided")
			}
			return family + ", " + given, nil
		},

		Success: func(data string, values map[string]string, w http.ResponseWriter, r *http.Request) error {
			return content.WriteHTML(data, nil, successTemplate, w, r)
		},
	}

	fmt.Println(makeFormRequest(nil, &form, nil))
	fmt.Println(makeFormRequest(nil, &form, map[string]string{"givenName": "Andrea", "familyName": "", "password": "something"}))
	fmt.Println(makeFormRequest(nil, &form, map[string]string{"givenName": "Andrea", "familyName": "Picard", "password": "something"}))

	// Output: "GET" returned code 200 with text/html; charset=utf-8 "<!doctype html><title>Form</title><input name=givenName placeholder>\n<input name=familyName placeholder>\n<input type=password name=password placeholder>"
	// "POST" returned code 200 with text/html; charset=utf-8 "<!doctype html><title>Form</title><p>Error: family name must not be empty</p><input value=Andrea name=givenName placeholder>\n<input name=familyName placeholder>\n<input type=password name=password placeholder>"
	// "POST" returned code 200 with text/html; charset=utf-8 "<!doctype html><title>Success</title>Welcome Picard, Andrea"
}

// makeFormRequest makes a request to a form
func makeFormRequest(t *testing.T, form http.Handler, body map[string]string) string {
	if t != nil {
		t.Helper()
	}

	var req *http.Request
	if body == nil {
		var err error
		req, err = http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			panic(err)
		}
	} else {
		form := url.Values{}
		for name, value := range body {
			form.Set(name, value)
		}

		var err error
		req, err = http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		if err != nil {
			panic(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	rr := httptest.NewRecorder()
	form.ServeHTTP(rr, req)

	rrr := rr.Result()
	result, _ := io.ReadAll(rrr.Body)
	return fmt.Sprintf("%q returned code %d with %s %q", req.Method, rrr.StatusCode, rrr.Header.Get("Content-Type"), string(result))
}
