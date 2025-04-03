//spellchecker:words reflectx
package reflectx

//spellchecker:words iter reflect slices
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

func iterFields(embeds bool, index []int, typ reflect.Type) iter.Seq2[reflect.StructField, []int] {
	return func(yield func(reflect.StructField, []int) bool) {
		for i := range typ.NumField() {
			field := typ.Field(i)
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
