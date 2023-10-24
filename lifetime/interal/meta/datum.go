package meta

import (
	"reflect"

	"github.com/tkw1536/pkglib/collection"
	"github.com/tkw1536/pkglib/reflectx"
)

// Datum holds information about a specific component.
type Datum[Component any] struct {
	Name string       // the type name of this component
	Elem reflect.Type // the element type of the component

	CFields map[string]reflect.Type // fields with type C for which C implements component
	IFields map[string]reflect.Type // fields []I where I is an interface that implements component

	DCFields map[string]reflect.Type // fields with type C for which C inside auto field which implement component
	DIFields map[string]reflect.Type // fields []I where I is an interface inside auto field that implements component
}

// newDatum creates a new datum by scanning for all relevant fields
func newDatum[Component any](component reflect.Type, concrete reflect.Type) (m Datum[Component]) {
	if concrete.Kind() != reflect.Pointer && concrete.Elem().Kind() != reflect.Struct {
		panic("newDatum: Type (" + concrete.String() + ") must be backed by a pointer to slice")
	}

	m.Elem = concrete.Elem()
	m.Name = reflectx.NameOf(m.Elem)

	m.CFields = make(map[string]reflect.Type)
	m.IFields = make(map[string]reflect.Type)
	scanForFields(component, m.Name, m.Elem, false, m.CFields, m.IFields)

	// check if we have a dependencies field of struct type
	dependenciesField, ok := m.Elem.FieldByName(dependencies)
	if !ok {
		return
	}

	if dependenciesField.Type.Kind() != reflect.Struct {
		panic("newDatum: " + dependencies + " field (" + m.Name + ") is not a struct")
	}

	// and initialize the type map of the given map
	m.DCFields = make(map[string]reflect.Type)
	m.DIFields = make(map[string]reflect.Type)
	scanForFields(component, m.Name, dependenciesField.Type, true, m.DCFields, m.DIFields)

	return
}

// scanForFields scans the struct type for fields of component-like fields.
// they are then written to the cFields and iFields maps.
// inDependenciesStruct indicates if we are inside a dependency struct
func scanForFields(component reflect.Type, elem string, structType reflect.Type, inDependenciesStruct bool, cFields map[string]reflect.Type, iFields map[string]reflect.Type) {
	count := structType.NumField()
	for i := 0; i < count; i++ {
		field := structType.Field(i)

		if !inDependenciesStruct && field.Tag.Get("auto") != "true" {
			continue
		}
		if !field.IsExported() {
			panic("newDatum: " + dependencies + " field (" + elem + ") contains field (" + field.Name + ") which is not exported")
		}
		if inDependenciesStruct && field.Tag != "" {
			panic("newDatum: " + dependencies + " field (" + elem + ") contains field (" + field.Name + ") with tag")
		}

		tp := field.Type
		name := field.Name

		switch {
		case implementsAsStructPointer(component, tp):
			cFields[name] = tp
		case implementsAsSliceInterface(component, tp):
			iFields[name] = tp.Elem()
		case inDependenciesStruct:
			panic("newDatum: " + dependencies + " field (" + elem + ") contains non-auto fields")
		}
	}
}

// New creates a new component of the concrete type this datum describes
func (m Datum[Component]) New() Component {
	return reflect.New(m.Elem).Interface().(Component)
}

// NeedsInitComponent checks if dependencies need to be injected into this component.
func (m Datum[Component]) NeedsInitComponent() bool {
	return len(m.CFields) > 0 || len(m.IFields) > 0 || len(m.DCFields) > 0 || len(m.DIFields) > 0
}

// name of the dependencies field
const dependencies = "Dependencies"

// InitComponent sets up the fields of the given instance of a component.
func (m Datum[Component]) InitComponent(instance reflect.Value, all []Component) {
	elem := instance.Elem()
	dependenciesElem := elem.FieldByName(dependencies)

	// assign the component fields
	for field, eType := range m.CFields {
		c := collection.First(all, func(c Component) bool {
			return reflect.TypeOf(c).AssignableTo(eType)
		})

		field := elem.FieldByName(field)
		unsafeSetAnyValue(field, reflect.ValueOf(c))
	}
	for field, eType := range m.DCFields {
		c := collection.First(all, func(c Component) bool {
			return reflect.TypeOf(c).AssignableTo(eType)
		})

		field := dependenciesElem.FieldByName(field)
		unsafeSetAnyValue(field, reflect.ValueOf(c))
	}

	// assign the interface subtypes
	registryR := reflect.ValueOf(all)
	for field, eType := range m.IFields {
		cs := filterSliceInterface(registryR, eType)
		field := elem.FieldByName(field)
		unsafeSetAnyValue(field, cs)
	}
	for field, eType := range m.DIFields {
		cs := filterSliceInterface(registryR, eType)
		field := dependenciesElem.FieldByName(field)
		unsafeSetAnyValue(field, cs)
	}
}
