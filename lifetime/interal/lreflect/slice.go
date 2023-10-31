package lreflect

import (
	"fmt"
	"reflect"
)

// ImplementsAsSliceInterface checks if T is a slice type with an interface element that implements I.
// I must be an interface, T may be any type.
func ImplementsAsSliceInterface(I reflect.Type, T reflect.Type) (bool, error) {

	// check for valid arguments
	{
		if T == nil {
			return false, errNilType("T")
		}
		if I == nil {
			return false, errNilType("I")
		}
		if I.Kind() != reflect.Interface {
			return false, errNoInterface("I")
		}
	}

	return T.Kind() == reflect.Slice && T.Elem().Kind() == reflect.Interface && T.Elem().Implements(I), nil
}

// FilterSliceInterface filters the slice S by all elements which implement the interface I and returns a new slice of I.
// slice must be a slice of some type (preferably some interface), I must be an interface.
func FilterSliceInterface(slice reflect.Value, I reflect.Type) (reflect.Value, error) {
	// check that we have valid arguments
	{
		S := slice.Type()
		if S == nil {
			return reflect.Value{}, errNilType("slice.Type()")
		}
		if S.Kind() != reflect.Slice {
			return reflect.Value{}, errNoSlice("slice.Type()")
		}
		if I == nil {
			return reflect.Value{}, errNilType("I")
		}
		if I.Kind() != reflect.Interface {
			return reflect.Value{}, errNoInterface("I")
		}
	}

	// create a new slice
	len := slice.Len()
	result := reflect.MakeSlice(reflect.SliceOf(I), 0, len)

	// iterate over the elements and check if they implement the slice
	for i := 0; i < len; i++ {
		element := slice.Index(i)
		if element.Elem().Type().Implements(I) {
			result = reflect.Append(result, element.Elem().Convert(I))
		}
	}

	return result, nil
}

// FirstAssignableElement finds the first element in slice that is assignable V.
// If no such element exists, returns the zero value of V.
//
// slice must be a slice of some interface type.
func FirstAssignableInterfaceElement(slice reflect.Value, V reflect.Type) (reflect.Value, error) {
	// check that we have valid arguments
	{
		S := slice.Type()
		if S == nil {
			return reflect.Value{}, errNilType("slice.Type()")
		}

		if S.Kind() != reflect.Slice || S.Elem().Kind() != reflect.Interface {
			return reflect.Value{}, errNoInterfaceSlice("slice.Type()")
		}

		if V == nil {
			return reflect.Value{}, errNilType("V")
		}
	}

	// find an element that is assignable to V
	len := slice.Len()
	for i := 0; i < len; i++ {
		element := slice.Index(i).Elem()
		if element.Type().AssignableTo(V) {
			return element, nil
		}
	}

	// no element found => return the nil value of V
	return reflect.New(V).Elem(), nil
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

// errNoSlice indicates that the type with the provided name is not a slice
type errNoSlice string

func (err errNoSlice) Error() string {
	return fmt.Sprintf("%s must be a slice type", string(err))
}

// errNilType indicates that the type with the provided name is nil
type errNilType string

func (err errNilType) Error() string {
	return fmt.Sprintf("%s must not be a nil type", string(err))
}

// errNoInterface indicates that the type with the provided name is not an interface
type errNoInterface string

func (err errNoInterface) Error() string {
	return fmt.Sprintf("%s must be an interface type", string(err))
}

// errNoInterfaceSlice indicates that the type with provided name is not a slice of an interface
type errNoInterfaceSlice string

func (err errNoInterfaceSlice) Error() string {
	return fmt.Sprintf("%s must be a slice of some interface type", string(err))
}
