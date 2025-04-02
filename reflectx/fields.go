package reflectx

import (
	"iter"
	"reflect"
	"slices"
)

// IterFields returns an iterator that iterates over the struct fields of typ in the order they are returned by [reflect.Field].
//
// When typ is not a struct type, IterFields panics.
// Unlike [IterAllFields], does not iterate over embedded fields recursively.
func IterFields(typ reflect.Type) iter.Seq2[reflect.StructField, int] {
	if typ.Kind() != reflect.Struct {
		panic("IterFields: typ is not a Struct")
	}

	return func(yield func(reflect.StructField, int) bool) {
		for field, indexes := range iterFields(false, nil, typ) {
			if !yield(field, indexes[0]) {
				return
			}
		}
	}
}

// IterateFields iterates over the struct fields of typ and calls f for each field.
// Does not recurse into embedded fields.
//
// The return value of f indicates if the iteration should be stopped early.
// Returns the last value returned by f, or false.
//
// Deprecated: Use [IterFields] instead.
func IterateFields(typ reflect.Type, f func(field reflect.StructField, index int) (stop bool)) (cancelled bool) {
	if typ.Kind() != reflect.Struct {
		panic("IterateFields: T is not a Struct")
	}

	for field, indexes := range iterFields(false, nil, typ) {
		if f(field, indexes[0]) {
			return true
		}
	}
	return false
}

// IterateAllFields iterates over the struct fields of typ and their indexes.
// Fields are iterated in the order they are returned by reflect.Field().
//
// When typ is not a struct type, IterAllFields panics.
// When typ contains an embedded struct, calls IterAllFields recursively.
//
// Unlike IterFields, this function recurses into embedded structs.
// See also IterFields.
func IterAllFields(typ reflect.Type) iter.Seq2[reflect.StructField, []int] {
	if typ.Kind() != reflect.Struct {
		panic("IterAllFields: typ is not a Struct")
	}

	return iterFields(true, nil, typ)
}

// IterateAllFields iterates over the struct fields of typ and calls f for each field.
//
// The return value of f indicates if the iteration should be stopped early.
// If f returns true, no further calls to f are made.
// IterateAllFields returns the return value of the last call to f, or false.
//
// Deprecated: Use [IterAllFields] instead.
func IterateAllFields(typ reflect.Type, f func(field reflect.StructField, index ...int) (stop bool)) (stopped bool) {
	if typ.Kind() != reflect.Struct {
		panic("IterateAllFields: typ is not a Struct")
	}

	for field, index := range iterFields(true, nil, typ) {
		if f(field, index...) {
			return true
		}
	}

	return false
}

func iterFields(embeds bool, index []int, T reflect.Type) iter.Seq2[reflect.StructField, []int] {
	return func(yield func(reflect.StructField, []int) bool) {
		for i := range T.NumField() {
			field := T.Field(i)
			fieldIndex := append(slices.Clone(index), i)
			if embeds && field.Anonymous && field.Type.Kind() == reflect.Struct {
				for field, indexes := range iterFields(embeds, fieldIndex, field.Type) {
					if !yield(field, indexes) {
						return
					}
				}
				continue
			}
			if !yield(field, fieldIndex) {
				return
			}
		}
	}

}
