//spellchecker:words field
package field

//spellchecker:words html template
import (
	"fmt"
	"html/template"
	"io"
)

// DefaultFieldTemplate is the default template to render fields.
var DefaultFieldTemplate = template.Must(template.New("").Parse(`
{{ if (eq .Type "textarea" ) }}
<textarea name="{{.Name}}" id="{{.Name}}" placeholder="{{.Placeholder}}" {{if .Autocomplete }} autocomplete="{{.Autocomplete}}" {{end}}>{{.Value}}</textarea>
{{ else }}
<input type="{{.Type}}" value="{{.Value}}" name="{{.Name}}" placeholder="{{.Placeholder}}" {{if .Autocomplete }} autocomplete="{{.Autocomplete}}" {{end}}>
{{ end }}`))

// Field represents a field inside a form.
type Field struct {
	Name string    // Name is the name of the field
	Type InputType // Type is the type of the field. It corresponds to the "name" attribute in html.

	Placeholder string // Value for the "placeholder" attribute
	Label       string // (External) Label for the field. Not used by the default template.

	Autocomplete Autocomplete

	EmptyOnError bool // indicates if the field should be reset on error
}

// fieldContext is passed to the template context.
type fieldContext struct {
	Field
	Value string
}

func (field Field) WriteTo(w io.Writer, template *template.Template, value string) error {
	if template == nil {
		template = DefaultFieldTemplate
	}
	err := template.Execute(w, fieldContext{Field: field, Value: value})
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}

// CheckboxChecked is the default value of a checked checkbox.
const CheckboxChecked = "on"
