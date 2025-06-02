package souls

import (
	"errors"
	"fmt"
	"iter"
	"reflect"

	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
	"github.com/tkw1536/pkglib/reflectx"
)

// dependency represents a single dependency.
type dependency struct {
	field reflect.StructField

	IsSlice bool

	Dependencies bool
}

// Get returns a reflect.Value pointing to the given field on the given component instance.
func (dep dependency) Get(instance reflect.Value) (reflect.Value, error) {
	if typ := instance.Type(); typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, errNotPointerToStruct
	}

	structValue := instance.Elem()
	if dep.Dependencies {
		field := structValue.FieldByName(dependenciesFieldName)
		if !field.IsValid() {
			return reflect.Value{}, fmt.Errorf("%q: %w", dependenciesFieldName, errNoSuchField)
		}
		if field.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("%q: %w", dependenciesFieldName, errNotAStruct)
		}
		structValue = field
	}

	value := structValue.FieldByIndex(dep.field.Index)
	if !value.IsValid() {
		return reflect.Value{}, fmt.Errorf("%q: %w", dep.field.Name, errNoSuchField)
	}
	return lreflect.UnsafeForgetUnexported(value, value.Type()), nil
}

// field name of this dependency.
func (dep dependency) Name() string {
	return dep.field.Name
}

// Element type of this dependency.
func (dep dependency) Elem() reflect.Type {
	if dep.IsSlice {
		return dep.field.Type.Elem()
	}
	return dep.field.Type
}

// Scan iterates through the dependencies of the given concrete and component types.
func Scan(component reflect.Type, concrete reflect.Type) iter.Seq2[dependency, error] {
	return func(yield func(dependency, error) bool) {
		// check that the types themselves are valid
		if component == nil || concrete == nil || concrete.Kind() != reflect.Pointer || concrete.Elem().Kind() != reflect.Struct {
			yield(dependency{}, newDepsError(dependency{}, concrete, errNotPointerToStruct))
			return
		}

		elem := concrete.Elem()
		for dep, err := range scan(component, concrete.Elem(), false) {
			if err != nil {
				yield(dep, newDepsError(dep, concrete, err))
				return
			}
			if !yield(dep, nil) {
				return
			}
		}

		dependenciesField, ok := elem.FieldByName(dependenciesFieldName)
		if !ok {
			return
		}

		if dependenciesField.Type.Kind() != reflect.Struct {
			yield(
				dependency{},
				newDepsError(dependency{field: dependenciesField, Dependencies: false}, concrete, errNotAStruct),
			)
			return
		}

		for dep, err := range scan(component, dependenciesField.Type, true) {
			if err != nil {
				yield(dep, newDepsError(dep, concrete, err))
				return
			}

			if !yield(dep, nil) {
				return
			}
		}
	}
}

func scan(component reflect.Type, typ reflect.Type, inDependenciesStruct bool) iter.Seq2[dependency, error] {
	return func(yield func(dependency, error) bool) {
		for field := range reflectx.IterFields(typ) {
			if !inDependenciesStruct && field.Tag.Get(injectFieldName) != "true" {
				continue
			}
			if inDependenciesStruct && field.Tag != "" {
				yield(
					dependency{
						field:        field,
						Dependencies: inDependenciesStruct,
					},
					errFieldHasTag,
				)
				return
			}

			tp := field.Type

			{
				isSingleComponent, err := lreflect.ImplementsAsStructPointer(component, tp)
				if err != nil {
					yield(
						dependency{
							field:        field,
							Dependencies: inDependenciesStruct,
						},
						err,
					)
					return
				}
				if isSingleComponent {
					if !yield(dependency{
						field:        field,
						IsSlice:      false,
						Dependencies: inDependenciesStruct,
					}, nil) {
						return
					}
					continue
				}
			}

			{
				isSubtype, err := lreflect.ImplementsAsSliceInterface(component, tp)
				if err != nil {
					yield(
						dependency{
							field:        field,
							Dependencies: inDependenciesStruct,
						},
						err,
					)
					return
				}
				if isSubtype {
					if !yield(dependency{
						field:        field,
						IsSlice:      true,
						Dependencies: inDependenciesStruct,
					}, nil) {
						return
					}
					continue
				}
			}

			if inDependenciesStruct {
				yield(
					dependency{
						field:        field,
						Dependencies: inDependenciesStruct,
					},
					errNotInjectField,
				)
				return
			}
		}
	}
}

var (
	errFieldHasTag    = errors.New("field has tag")
	errNotInjectField = errors.New("not an injected field")

	errNotPointerToStruct      = errors.New("not a pointer to a slice")
	errNotAStruct              = errors.New("not a struct")
	errComponentNotImplemented = errors.New("type does not implement component")

	errNoSuchField = errors.New("no such field")
)

func newDepsError(dep dependency, concrete reflect.Type, err error) error {
	name := dep.Name()
	if name == "" {
		return fmt.Errorf("type %s: %w", concrete, err)
	}

	var fieldPrefix string
	if dep.Dependencies {
		fieldPrefix = "dependencies "
	}
	return fmt.Errorf("type %s, %sfield %q: %w", concrete, fieldPrefix, name, err)
}
