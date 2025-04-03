//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import (
	"fmt"
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
			return invalidValueError("v")
		}
		if !x.IsValid() {
			return invalidValueError("x")
		}
	}

	// ensure that the types are
	xT := x.Type()
	vT := v.Type()
	if !xT.AssignableTo(vT) {
		return typeUnassignableError{X: xT, V: vT}
	}

	// check if the value was obtained from an unexported field
	// and "forget" where the value was obtained
	if !v.CanSet() {
		if !v.CanAddr() {
			return typeUnaddressableError{X: vT}
		}

		v = reflect.NewAt(vT, v.Addr().UnsafePointer()).Elem()
	}

	// do the actual setting!
	v.Set(x)
	return nil
}

type typeUnassignableError struct {
	X, V reflect.Type
}

func (err typeUnassignableError) Error() string {
	return fmt.Sprintf("value of type %s not assignable to type %s", err.X, err.V)
}

type typeUnaddressableError struct {
	X reflect.Type
}

func (err typeUnaddressableError) Error() string {
	return fmt.Sprintf("value of type %s is not addressable", err.X)
}

// invalidValueError indicates that an invalid value was passed for the variable with the given name.
type invalidValueError string

func (err invalidValueError) Error() string {
	return string(err) + " is not a valid value"
}
