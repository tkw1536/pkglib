// Package reflectx provides extensions to the reflect package
//
//spellchecker:words reflectx
package reflectx

//spellchecker:words reflect slices
import (
	"reflect"
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
