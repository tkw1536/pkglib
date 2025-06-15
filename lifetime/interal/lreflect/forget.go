//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import "reflect"

// UnsafeForgetUnexported returns the same value as v, but forgets it was retrieved from an unexported field.
// vT should be the type of v.
//
// DO NOT USE THIS UNLESS YOU KNOW WHAT YOU'RE DOING.
func UnsafeForgetUnexported(v reflect.Value, vT reflect.Type) reflect.Value {
	return reflect.NewAt(vT, v.Addr().UnsafePointer()).Elem()
}
