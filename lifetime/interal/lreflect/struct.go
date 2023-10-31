// Package lreflect provides reflect extensions for use by lifetime
package lreflect

// spellchecker:words lreflect

import (
	"reflect"
)

// This file contains any methods directly interacting with methods via reflect.

// ImplementsStructAsPointer checks if T implements I and T is a pointer to a struct.
// I must be an interface type, T may be any type.
func ImplementsAsStructPointer(I reflect.Type, T reflect.Type) (bool, error) {
	{
		if I == nil {
			return false, errNilType("I")
		}
		if I.Kind() != reflect.Interface {
			return false, errNoInterface("I")
		}
		if T == nil {
			return false, errNilType("T")
		}
	}

	return T.Implements(I) && T.Kind() == reflect.Pointer && T.Elem().Kind() == reflect.Struct, nil
}
