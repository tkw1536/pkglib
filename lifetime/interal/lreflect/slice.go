//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import (
	"fmt"
	"reflect"
)

//spellchecker:words iface

// ImplementsAsSliceInterface checks if slice is a slice type with an interface element that implements iface.
// I must be an interface, T may be any type.
func ImplementsAsSliceInterface(iface reflect.Type, slice reflect.Type) (bool, error) {

	// check for valid arguments
	{
		if slice == nil {
			return false, nilTypeError("T")
		}
		if iface == nil {
			return false, nilTypeError("iface")
		}
		if iface.Kind() != reflect.Interface {
			return false, noInterfaceError("iface")
		}
	}

	return slice.Kind() == reflect.Slice && slice.Elem().Kind() == reflect.Interface && slice.Elem().Implements(iface), nil
}

// FilterSliceInterface filters the slice S by all elements which implement the interface iface and returns a new slice of I.
// slice must be a slice of some type (preferably some interface), I must be an interface.
func FilterSliceInterface(slice reflect.Value, iface reflect.Type) (reflect.Value, error) {
	// check that we have valid arguments
	{
		S := slice.Type()
		if S.Kind() != reflect.Slice {
			return reflect.Value{}, noSliceError("slice.Type()")
		}
		if iface == nil {
			return reflect.Value{}, nilTypeError("I")
		}
		if iface.Kind() != reflect.Interface {
			return reflect.Value{}, noInterfaceError("I")
		}
	}

	// create a new slice
	len := slice.Len()
	result := reflect.MakeSlice(reflect.SliceOf(iface), 0, len)

	// iterate over the elements and check if they implement the slice
	for i := range len {
		element := slice.Index(i)
		if element.Elem().Type().Implements(iface) {
			result = reflect.Append(result, element.Elem().Convert(iface))
		}
	}

	return result, nil
}

// FirstAssignableElement finds the first element in slice that is assignable to v.
// If no such element exists, returns the zero value of v.
//
// slice must be a slice of some interface type.
func FirstAssignableInterfaceElement(slice reflect.Value, v reflect.Type) (reflect.Value, error) {
	// check that we have valid arguments
	{
		s := slice.Type()

		if s.Kind() != reflect.Slice || s.Elem().Kind() != reflect.Interface {
			return reflect.Value{}, noInterfaceSliceError("slice.Type()")
		}

		if v == nil {
			return reflect.Value{}, nilTypeError("V")
		}
	}

	// find an element that is assignable to V
	len := slice.Len()
	for i := range len {
		element := slice.Index(i).Elem()
		if element.Type().AssignableTo(v) {
			return element, nil
		}
	}

	// no element found => return the nil value of V
	return reflect.New(v).Elem(), nil
}

// CopySlice makes a copy of the provided slice.
// When slice is not a slice, the behavior is undefined.
func CopySlice(slice reflect.Value) reflect.Value {
	if !slice.IsValid() || slice.Kind() != reflect.Slice {
		return reflect.Value{}
	}

	// if the passed slice is nil, return a new nil
	if slice.IsNil() {
		return reflect.New(slice.Type()).Elem()
	}

	// create a new slice and copy over the elements
	copy := reflect.MakeSlice(slice.Type(), slice.Len(), slice.Len())
	reflect.Copy(copy, slice)
	return copy
}

// noSliceError indicates that the type with the provided name is not a slice
type noSliceError string

func (err noSliceError) Error() string {
	return fmt.Sprintf("%s must be a slice type", string(err))
}

// nilTypeError indicates that the type with the provided name is nil
type nilTypeError string

func (err nilTypeError) Error() string {
	return fmt.Sprintf("%s must not be a nil type", string(err))
}

// noInterfaceError indicates that the type with the provided name is not an interface
type noInterfaceError string

func (err noInterfaceError) Error() string {
	return fmt.Sprintf("%s must be an interface type", string(err))
}

// noInterfaceSliceError indicates that the type with provided name is not a slice of an interface
type noInterfaceSliceError string

func (err noInterfaceSliceError) Error() string {
	return fmt.Sprintf("%s must be a slice of some interface type", string(err))
}
