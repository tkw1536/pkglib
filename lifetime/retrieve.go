package lifetime

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tkw1536/pkglib/lifetime/interal/souls"
)

var (
	errNotFound       = errors.New("direct dependency not found")
	errNotImplemented = errors.New("ConcreteComponentType not implemented by Component")
)

// Retrieve retrieves a declared (non-transitive) dependency of ConcreteComponentType from the given component instance.
// instance is assumed to have been initialized by a lifetime.
func Retrieve[ConcreteComponentType any, Component any](instance Component) (concrete ConcreteComponentType, err error) {
	// check that ConcreteComponentType implements Component
	// and that Component is an interface
	componentTyp := reflect.TypeFor[Component]()
	concreteType := reflect.TypeFor[ConcreteComponentType]()
	if componentTyp == nil || componentTyp.Kind() != reflect.Interface || !concreteType.Implements(componentTyp) {
		return concrete, errNotImplemented
	}

	instanceValue := reflect.ValueOf(instance)

	// TODO: Validate ConcreteComponentType

	for dep, err := range souls.Scan(reflect.TypeFor[Component](), instanceValue.Type()) {
		if err != nil {
			return concrete, err
		}
		if dep.IsSlice {
			continue
		}
		if dep.Elem() != concreteType {
			continue
		}

		element, err := dep.Get(instanceValue)
		if err != nil {
			return concrete, fmt.Errorf("failed to get value: %w", err)
		}
		return element.Interface().(ConcreteComponentType), nil
	}

	return concrete, errNotFound
}

// RetrieveSlice extracts a declared (non-transitive) slice dependency of []ConcreteComponentType from the given instance.
// instance is assumed to have been initialized by a lifetime.
func RetrieveSlice[ConcreteComponentType any, Component any](instance Component) (slice []ConcreteComponentType, err error) {
	// check that ConcreteComponentType implements Component
	// and that Component is an interface
	componentTyp := reflect.TypeFor[Component]()
	concreteType := reflect.TypeFor[ConcreteComponentType]()
	if componentTyp == nil || componentTyp.Kind() != reflect.Interface || !concreteType.Implements(componentTyp) {
		return nil, errNotImplemented
	}

	instanceValue := reflect.ValueOf(instance)

	for dep, err := range souls.Scan(reflect.TypeFor[Component](), instanceValue.Type()) {
		if err != nil {
			return nil, err
		}
		if !dep.IsSlice {
			continue
		}
		if dep.Elem() != concreteType {
			continue
		}

		element, err := dep.Get(instanceValue)
		if err != nil {
			return nil, fmt.Errorf("failed to get value: %w", err)
		}
		return element.Interface().([]ConcreteComponentType), nil
	}

	return nil, errNotFound
}
