//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import (
	"reflect"
)

//spellchecker:words unassignable unaddressable

// UnsafeSetAnyValue is like v.Set(x) except that it permits a value obtained from an unexported field to be set.
// It never panics, and instead returns an error.
//
// DO NOT USE THIS UNLESS YOU KNOW WHAT YOU'RE DOING.
func UnsafeSetAnyValue(v, x reflect.Value) error {
	// ensure both arguments are valid
	{
		if !v.IsValid() {
			return errVIsInvalidValue
		}
		if !x.IsValid() {
			return errXIsInvalidValue
		}
	}

	// ensure that the types are
	xT := x.Type()
	vT := v.Type()
	if !xT.AssignableTo(vT) {
		return typeUnassignableError{X: xT, V: vT}
	}

	// if we can directly set the value, do it!
	if v.CanSet() {
		v.Set(x)
		return nil
	}

	// ensure that we can address v!
	if !v.CanAddr() {
		return typeUnaddressableError{X: vT}
	}

	// forget that the value was retrieved from an unexported field
	// and immediately set it!
	UnsafeForgetUnexported(v, vT).Set(x)
	return nil
}
