//spellchecker:words souls
package souls

//spellchecker:words errors reflect github pkglib lifetime interal lreflect
import (
	"reflect"
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
	s.Elem = concrete.Elem()

	s.CFields = make(map[string]reflect.Type)
	s.IFields = make(map[string]reflect.Type)

	s.DCFields = make(map[string]reflect.Type)
	s.DIFields = make(map[string]reflect.Type)

	for dep, err := range Scan(component, concrete) {
		if err != nil {
			return soul{}, err
		}

		name := dep.Name()
		elem := dep.Elem()
		switch {
		case !dep.Dependencies && !dep.IsSlice:
			s.CFields[name] = elem
		case !dep.Dependencies && dep.IsSlice:
			s.IFields[name] = elem
		case dep.Dependencies && !dep.IsSlice:
			s.DCFields[name] = elem
		case dep.Dependencies && dep.IsSlice:
			s.DIFields[name] = elem
		}
	}

	return s, nil
}
