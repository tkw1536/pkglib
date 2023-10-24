package meta

import (
	"reflect"
)

// This file contains any methods directly interacting with methods via reflect.

// unsafeSetAnyValue is like v.Set(x) except that it permits a value obtained from an unexported field to be set.
// DO NOT USE THIS UNLESS YOU KNOW WHAT YOU'RE DOING.
func unsafeSetAnyValue(v, x reflect.Value) {
	// check if the value was obtained from an unexported field
	// and "forget" where the value was obtained
	if !v.CanSet() && v.CanAddr() {
		v = reflect.NewAt(v.Type(), v.Addr().UnsafePointer()).Elem()
	}

	// do the actual setting!
	v.Set(x)
}

// implementsAsStructPointer checks if T implements I and T is a pointer to a struct.
// I must be an interface type, T may be any type.
func implementsAsStructPointer(I reflect.Type, T reflect.Type) bool {
	if T == nil || I == nil || I.Kind() != reflect.Interface {
		return false
	}

	return T.Implements(I) && T.Kind() == reflect.Pointer && T.Elem().Kind() == reflect.Struct
}

// implementsAsSliceInterface checks if T is a slice type with an interface element that implements I.
// I must be an interface, T may be any type.
func implementsAsSliceInterface(I reflect.Type, T reflect.Type) bool {
	if T == nil || I == nil || I.Kind() != reflect.Interface {
		return false
	}
	return T.Kind() == reflect.Slice && T.Elem().Kind() == reflect.Interface && T.Elem().Implements(I)
}

// filterSliceInterface filters the slice S by all elements which implement the interface I and returns a new slice of I.
// slice must be a slice of some type (preferably some interface), I must be an interface.
func filterSliceInterface(slice reflect.Value, I reflect.Type) reflect.Value {
	// check that we have valid arguments
	if S := slice.Type(); S == nil || S.Kind() != reflect.Slice || I.Kind() != reflect.Interface {
		// don't have valid arguments => return a nil slice
		return reflect.New(reflect.SliceOf(I)).Elem()
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

	return result

}
