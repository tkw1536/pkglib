//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect
import (
	"fmt"
	"reflect"
)

//spellchecker:words iface unaddressable

const (
	errIfaceIsNilType  = nilTypeError("iface")
	errIfaceNotAnIface = noInterfaceError("iface")

	errSliceInvalid   = invalidValueError("slice")
	errSliceNotASlice = noSliceError("slice")

	errSliceTypeNotAnInterfaceSlice = noInterfaceSliceError("slice.Type()")
	errSliceTypeNotASlice           = noSliceError("slice.Type()")
	errSliceTypeIsNilType           = nilTypeError("slice")

	errSPtrIsNilType = nilTypeError("sPtr")

	errVIsInvalidValue = invalidValueError("v")
	errVIsNilType      = nilTypeError("v")

	errXIsInvalidValue = invalidValueError("x")
)

// noSliceError indicates that the type with the provided name is not a slice.
type noSliceError string

func (err noSliceError) Error() string {
	return string(err) + " must be a slice type"
}

// nilTypeError indicates that the type with the provided name is nil.
type nilTypeError string

func (err nilTypeError) Error() string {
	return string(err) + " must not be a nil type"
}

// noInterfaceError indicates that the type with the provided name is not an interface.
type noInterfaceError string

func (err noInterfaceError) Error() string {
	return string(err) + " must be an interface type"
}

// noInterfaceSliceError indicates that the type with provided name is not a slice of an interface.
type noInterfaceSliceError string

func (err noInterfaceSliceError) Error() string {
	return string(err) + " must be a slice of some interface type"
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
