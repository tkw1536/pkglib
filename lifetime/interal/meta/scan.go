package meta

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
)

// ComponentInfo holds information about a specific component
type ComponentInfo struct {
	Elem reflect.Type // the element type of the component

	CFields map[string]reflect.Type // fields with type C for which C implements component
	IFields map[string]reflect.Type // fields []I where I is an interface that implements component

	DCFields map[string]reflect.Type // fields with type C for which C inside auto field which implement component
	DIFields map[string]reflect.Type // fields []I where I is an interface inside auto field that implements component
}

// DependenciesField is the name of the dependencies field
const dependenciesFieldName = "dependencies"
const injectFieldName = "inject"

// New creates a new datum for the given concrete component
func New(component reflect.Type, concrete reflect.Type) (m ComponentInfo, err error) {
	if component == nil || concrete == nil || concrete.Kind() != reflect.Pointer || concrete.Elem().Kind() != reflect.Struct {
		return m, errNewGeneric{Concrete: concrete, Err: errNotPointerToStruct}
	}

	m.Elem = concrete.Elem()

	m.CFields = make(map[string]reflect.Type)
	m.IFields = make(map[string]reflect.Type)
	if err := m.scanFields(component, m.Elem, false, m.CFields, m.IFields); err != nil {
		return m, err
	}

	// check if we have a dependencies field of struct type
	dependenciesField, ok := m.Elem.FieldByName(dependenciesFieldName)
	if !ok {
		return
	}

	// check that we have a struct field
	if dependenciesField.Type.Kind() != reflect.Struct {
		return m, errDatumField{Concrete: concrete, InDependencies: false, Field: dependenciesFieldName, Err: errNotAStruct}
	}

	// and initialize the type map of the given map
	m.DCFields = make(map[string]reflect.Type)
	m.DIFields = make(map[string]reflect.Type)
	if err := m.scanFields(component, dependenciesField.Type, true, m.DCFields, m.DIFields); err != nil {
		return m, err
	}

	return
}

var errFieldHasTag = errors.New("field has tag")
var errNotInjectField = errors.New("not an injected field")

// scanFields scans the struct type for fields of component-like fields.
// they are then written to the cFields and iFields maps.
// inDependenciesStruct indicates if we are inside a dependency struct
func (m ComponentInfo) scanFields(component reflect.Type, structType reflect.Type, inDependenciesStruct bool, cFields map[string]reflect.Type, iFields map[string]reflect.Type) error {
	count := structType.NumField()
	for i := 0; i < count; i++ {
		field := structType.Field(i)

		if !inDependenciesStruct && field.Tag.Get(injectFieldName) != "true" {
			continue
		}
		if inDependenciesStruct && field.Tag != "" {
			return errDatumField{Concrete: m.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: errFieldHasTag}
		}

		tp := field.Type
		name := field.Name

		{
			isSingleComponent, err := lreflect.ImplementsAsStructPointer(component, tp)
			if err != nil {
				return errDatumField{Concrete: m.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: err}
			}
			if isSingleComponent {
				cFields[name] = tp
				continue
			}
		}

		{
			isSubtype, err := lreflect.ImplementsAsSliceInterface(component, tp)
			if err != nil {
				return errDatumField{Concrete: m.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: err}
			}
			if isSubtype {
				iFields[name] = tp.Elem()
				continue
			}
		}

		if inDependenciesStruct {
			return errDatumField{Concrete: m.Elem, InDependencies: inDependenciesStruct, Field: field.Name, Err: errNotInjectField}
		}
	}
	return nil
}

type errNewGeneric struct {
	Concrete reflect.Type
	Err      error
}

func (err errNewGeneric) Error() string {
	return fmt.Sprintf("New: Type %s: %s", err.Concrete, err.Err)
}

func (err errNewGeneric) Unwrap() error {
	return err.Err
}

type errDatumField struct {
	Concrete       reflect.Type
	InDependencies bool
	Field          string
	Err            error
}

func (err errDatumField) Error() string {
	var fieldPrefix string
	if err.InDependencies {
		fieldPrefix = "dependencies "
	}
	return fmt.Sprintf("New: Type %s, %sField %s: %s", err.Concrete, fieldPrefix, err.Field, err.Err)
}

func (err errDatumField) Unwrap() error {
	return err.Err
}

var (
	errNotPointerToStruct      = errors.New("not a pointer to a slice")
	errNotAStruct              = errors.New("not a struct")
	errComponentNotImplemented = errors.New("type does not implement component")
)
