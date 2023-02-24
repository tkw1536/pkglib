// Package reflectx provides extensions to the reflect package
package reflectx

import (
	"reflect"

	"golang.org/x/exp/slices"
)

// TypeOf returns the reflection Type that represents T.
//
// Unlike reflect.TypeOf, this method does not require instantiating T.
func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// MakePointerCopy returns a new copy of value of interface type I that is backed by a pointer (if possible).
// ptr holds the new copy, ok indiciates if the value is indeed a pointer.
func MakePointerCopy[I any](value I) (ptr I, ok bool) {
	iType := TypeOf[I]()
	if iType.Kind() != reflect.Interface {
		panic("MakePointer: I must be an interface type")
	}

	// get concrete type and value
	rValue := reflect.ValueOf(value)
	rTyp := rValue.Type()

	// check that the target type implements type
	tTyp := reflect.PointerTo(rTyp)
	if !tTyp.Implements(iType) {
		return Copy(value, true), rTyp.Kind() == reflect.Pointer
	}

	// easy case: we can address the value
	if rValue.CanAddr() {
		return rValue.Addr().Interface().(I), true // ok because of implements test above!
	}

	// create a new copy of value
	copy := reflect.New(rTyp)
	copy.Elem().Set(rValue)

	// safe because above
	return copy.Interface().(I), true
}

// Copy returns a copy of value.
// When copyElem is true, and value is of kind pointer, returns a pointer to a copy of the pointed to value.
func Copy[T any](value T, copyElem bool) T {
	rValue := reflect.ValueOf(value)
	rTyp := rValue.Type()

	// copy the underlying element if requested
	if copyElem && rTyp.Kind() == reflect.Pointer {
		eValue := rValue.Elem()
		eTyp := eValue.Type()

		copy := reflect.New(eTyp)
		copy.Elem().Set(eValue)

		return copy.Interface().(T)
	}

	// make a copy
	copy := reflect.New(rTyp)
	copy.Elem().Set(rValue)
	return copy.Elem().Interface().(T)
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
		panic("IterateFields: tp is not a Struct")
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
	nf := T.NumField()
	for i := 0; i < nf; i++ {
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
