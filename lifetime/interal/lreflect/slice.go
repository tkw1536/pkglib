//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import (
	"reflect"
)

//spellchecker:words iface

// ImplementsAsSliceInterface checks if slice is a slice type with an interface element that implements iface.
// iface must be an interface, slice may be any type.
func ImplementsAsSliceInterface(iface reflect.Type, slice reflect.Type) (bool, error) {
	// check for valid arguments
	{
		if slice == nil {
			return false, sliceIsNilTypeErr
		}
		if iface == nil {
			return false, ifaceIsNilTypeErr
		}
		if iface.Kind() != reflect.Interface {
			return false, ifaceNotAnIfaceErr
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
			return reflect.Value{}, sliceTypeNotASlice
		}
		if iface == nil {
			return reflect.Value{}, ifaceIsNilTypeErr
		}
		if iface.Kind() != reflect.Interface {
			return reflect.Value{}, ifaceNotAnIfaceErr
		}
	}

	// create a new slice
	sliceLen := slice.Len()
	result := reflect.MakeSlice(reflect.SliceOf(iface), 0, sliceLen)

	// iterate over the elements and check if they implement the slice
	for i := range sliceLen {
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
			return reflect.Value{}, sliceTypeNotAnInterfaceSlice
		}

		if v == nil {
			return reflect.Value{}, vIsNilTypeErr
		}
	}

	// find an element that is assignable to V
	for i := range slice.Len() {
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

	// create a new slice and cp over the elements
	cp := reflect.MakeSlice(slice.Type(), slice.Len(), slice.Len())
	reflect.Copy(cp, slice)
	return cp
}
