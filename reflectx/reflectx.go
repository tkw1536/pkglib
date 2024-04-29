// Package reflectx provides extensions to the reflect package
package reflectx

import (
	"reflect"
	"slices"
)

// CopyInterface returns a new copy of the interface value I.
// I must be an interface type.
//
// value is a copy of value, mutable indicates if the value is mutable, i.e. if it is backed by a pointer type.
//
// If value is backed by a pointer type, the pointed-to-value is copied, and a pointer to that copy and the boolean true is returned.
// If the value is not backed by a non-pointer type C, and *C implements I, then a pointer to the copy of the underlying value, along with the boolean true, is returned.
// Otherwise, a simple copy, and the boolean false, is returned.
func CopyInterface[I any](value I) (ptr I, mutable bool) {

	// ensure that we are dealing with an interface
	iTyp := reflect.TypeFor[I]()
	if iTyp.Kind() != reflect.Interface {
		panic("CopyInterface: I must be an interface type")
	}

	// get C, the concrete type backing value
	cTyp := reflect.TypeOf(value)

	// case 1: we have a pointer => copy the underlying value
	if cTyp.Kind() == reflect.Pointer {
		copy := reflect.New(cTyp.Elem())
		copy.Elem().Set(reflect.ValueOf(value).Elem())
		return copy.Interface().(I), true
	}

	// case 2: *C implements I => return a pointer to the copy
	if reflect.PointerTo(cTyp).Implements(iTyp) {
		copy := reflect.New(cTyp)
		copy.Elem().Set(reflect.ValueOf(value))
		return copy.Interface().(I), true
	}

	// case 3: *C does not implement I => fallback
	return value, false
}

// IterateFields iterates over the struct fields of T and calls f for each field.
// Fields are iterated in the order they are returned by reflect.Field().
//
// When T is not a struct type, IterateFields panics.
// Unlike IterateAllFields, does not iterate over embedded fields recursively.
//
// The return value of f indicates if the iteration should be cancelled early.
// If f returns true, no further calls to f are made.
//
// IterateFields returns if the iteration was aborted early.
//
// Unlike IterateAllFields, this function does not recurse into embedded structs.
// See also IterateFields.
func IterateFields(T reflect.Type, f func(field reflect.StructField, index int) (stop bool)) (cancelled bool) {
	if T.Kind() != reflect.Struct {
		panic("IterateFields: T is not a Struct")
	}

	return iterateFields(false, nil, T, func(field reflect.StructField, index ...int) (stop bool) {
		return f(field, index[0])
	})
}

// IterateAllFields iterates over the struct fields of T and calls f for each field.
// Fields are iterated in the order they are returned by reflect.Field().
//
// When T is not a struct type, IterateAllFields panics.
// When T contains an embedded struct, calls IterateAllFields recursively.
//
// The return value of f indicates if the iteration should be stopped early.
// If f returns true, no further calls to f are made.
//
// IterateAllFields returns if the iteration was aborted early.
//
// Unlike IterateFields, this function recurses into embedded structs.
// See also IterateFields.
func IterateAllFields(T reflect.Type, f func(field reflect.StructField, index ...int) (stop bool)) (stopped bool) {
	if T.Kind() != reflect.Struct {
		panic("IterateAllFields: tp is not a Struct")
	}

	return iterateFields(true, nil, T, f)
}

// iterateFields implements IterateFields and IterateAllFields
func iterateFields(embeds bool, index []int, T reflect.Type, f func(field reflect.StructField, index ...int) (cancel bool)) (cancelled bool) {
	for i := range T.NumField() {
		field := T.Field(i)
		fieldIndex := append(slices.Clone(index), i)
		if embeds && field.Anonymous && field.Type.Kind() == reflect.Struct {
			if iterateFields(embeds, fieldIndex, field.Type, f) {
				return true
			}
			continue
		}
		if f(field, fieldIndex...) {
			return true
		}
	}
	return false
}

// NameOf returns the fully qualified name for T.
//
// A fully qualified name consists of the package path, followed by a ".", followed by the type name.
// Builtin types have the empty package path.
// Types that are not named, return the empty string
func NameOf(T reflect.Type) string {
	if T == nil {
		return ""
	}

	name := T.Name()
	if name == "" {
		return ""
	}
	return T.PkgPath() + "." + name
}
