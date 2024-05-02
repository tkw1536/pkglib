// Package recover provides [Recover] and [Safe]
//
//spellchecker:words recovery
package recovery

//spellchecker:words runtime debug
import (
	"fmt"
	"runtime/debug"
)

// Recover returns an error that represents an error caught from recover.
// When passed nil, returns nil.
//
// It should be used as:
//
//	if err := Recover(recover()); err != nil {
//		// ... handle here ...
//	}
func Recover(value any) error {
	if value == nil {
		return nil
	}

	// TODO: build a custom stack trace here.
	// That way we can skip the current caller.
	// But that's too much effort for now.
	return recovered{
		Stack: debug.Stack(),
		Value: value,
	}
}

// Safe calls and returns the value of f.
// If a panic occurs, t is set to the zero value, and error as returned by [Recover].
func Safe[T any](f func() (T, error)) (t T, err error) {
	defer func() {
		// recover and replace any error
		if e := Recover(recover()); e != nil {
			err = e
		}
	}()

	return f()
}

// Safe2 is like Safe, but takes a function returning two values.
func Safe2[S, T any](f func() (S, T, error)) (s S, t T, err error) {
	defer func() {
		// recover and replace any error
		if e := Recover(recover()); e != nil {
			err = e
		}
	}()
	return f()
}

type recovered struct {
	Stack []byte
	Value any
}

func (r recovered) GoString() string {
	return fmt.Sprintf("recovery.recovered{/* recover() = %#v */}", r.Value)
}

func (r recovered) Error() string {
	return fmt.Sprintf("%v\n\n%s", r.Value, r.Stack)
}
