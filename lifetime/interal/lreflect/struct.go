// Package lreflect provides reflect extensions for use by the lifetime package.
//
//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import (
	"reflect"
)

//spellchecker:words iface

// This file contains any methods directly interacting with methods via reflect.

// ImplementsStructAsPointer checks if sPtr implements iface and sPtr is a pointer to a struct.
// iface must be an interface type, sPtr may be any type.
func ImplementsAsStructPointer(iface reflect.Type, sPtr reflect.Type) (bool, error) {
	{
		if iface == nil {
			return false, ifaceIsNilTypeErr
		}
		if iface.Kind() != reflect.Interface {
			return false, ifaceNotAnIfaceErr
		}
		if sPtr == nil {
			return false, sPtrNilErr
		}
	}

	return sPtr.Implements(iface) && sPtr.Kind() == reflect.Pointer && sPtr.Elem().Kind() == reflect.Struct, nil
}
