//spellchecker:words souls
package souls

//spellchecker:words errors reflect github pkglib lifetime interal lreflect
import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
)

// soul holds information about a specific component.
type soul struct {
	Elem reflect.Type // the element type of the component

	CFields map[string]reflect.Type // fields with type C for which C implements component
	IFields map[string]reflect.Type // fields []I where I is an interface that implements component

	DCFields map[string]reflect.Type // fields with type C for which C inside auto field which implement component
	DIFields map[string]reflect.Type // fields []I where I is an interface inside auto field that implements component
}

// DependenciesField is the name of the dependencies field.
const dependenciesFieldName = "dependencies"
const injectFieldName = "inject"

// newSoul creates a soul for the given concrete component.
func newSoul(component reflect.Type, concrete reflect.Type) (s soul, err error) {
	if component == nil || concrete == nil || concrete.Kind() != reflect.Pointer || concrete.Elem().Kind() != reflect.Struct {
		return s, newGenericError{Concrete: concrete, Err: errNotPointerToStruct}
	}

	s.Elem = concrete.Elem()

	s.CFields = make(map[string]reflect.Type)
	s.IFields = make(map[string]reflect.Type)
	if err := s.scanFields(component, s.Elem, false, s.CFields, s.IFields); err != nil {
		return s, err
	}

	// check if we have a dependencies field of struct type
	dependenciesField, ok := s.Elem.FieldByName(dependenciesFieldName)
	if !ok {
		return
	}

	// check that we have a struct field
	if dependenciesField.Type.Kind() != reflect.Struct {
		return s, datumFieldError{Concrete: concrete, InDependencies: false, Field: dependenciesFieldName, Err: errNotAStruct}
	}

	// and initialize the type map of the given map
	s.DCFields = make(map[string]reflect.Type)
	s.DIFields = make(map[string]reflect.Type)
	if err := s.scanFields(component, dependenciesField.Type, true, s.DCFields, s.DIFields); err != nil {
		return s, err
	}

	return
}

var errFieldHasTag = errors.New("field has tag")
var errNotInjectField = errors.New("not an injected field")

// inDependenciesStruct indicates if we are inside a dependency struct.
func (s soul) scanFields(component reflect.Type, structType reflect.Type, inDependenciesStruct bool, cFields map[string]reflect.Type, iFields map[string]reflect.Type) error {
	for i := range structType.NumField() {
		field := structType.Field(i)

		if !inDependenciesStruct && field.Tag.Get(injectFieldName) != "true" {
			continue
		}
		if inDependenciesStruct && field.Tag != "" {
			return datumFieldError{Concrete: s.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: errFieldHasTag}
		}

		tp := field.Type
		name := field.Name

		{
			isSingleComponent, err := lreflect.ImplementsAsStructPointer(component, tp)
			if err != nil {
				return datumFieldError{Concrete: s.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: err}
			}
			if isSingleComponent {
				cFields[name] = tp
				continue
			}
		}

		{
			isSubtype, err := lreflect.ImplementsAsSliceInterface(component, tp)
			if err != nil {
				return datumFieldError{Concrete: s.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: err}
			}
			if isSubtype {
				iFields[name] = tp.Elem()
				continue
			}
		}

		if inDependenciesStruct {
			return datumFieldError{Concrete: s.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: errNotInjectField}
		}
	}
	return nil
}

type newGenericError struct {
	Concrete reflect.Type
	Err      error
}

func (err newGenericError) Error() string {
	return fmt.Sprintf("New: Type %s: %s", err.Concrete, err.Err)
}

func (err newGenericError) Unwrap() error {
	return err.Err
}

type datumFieldError struct {
	Concrete       reflect.Type
	InDependencies bool
	Field          string
	Err            error
}

func (err datumFieldError) Error() string {
	var fieldPrefix string
	if err.InDependencies {
		fieldPrefix = "dependencies "
	}
	return fmt.Sprintf("New: Type %s, %sField %s: %s", err.Concrete, fieldPrefix, err.Field, err.Err)
}

func (err datumFieldError) Unwrap() error {
	return err.Err
}

var (
	errNotPointerToStruct      = errors.New("not a pointer to a slice")
	errNotAStruct              = errors.New("not a struct")
	errComponentNotImplemented = errors.New("type does not implement component")
)
