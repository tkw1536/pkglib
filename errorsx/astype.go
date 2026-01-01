package errorsx

import "errors"

// AsType is like [errors.As], except that it returns the type error and a boolean.
//
// A more efficient implementation of this function will be included in the go1.26 standard library.
// This current implementation will be removed in go1.26.
func AsType[E error](err error) (E, bool) {
	var (
		anError E
		// [errors.As] internally uses reflection.
		// This is technically inefficient.
		// We could backport the 1.26 standard library implementation now, but that would cause licensing issues.
		// So we don't bother with it; as what callers do now is equivalent to this.
		ok = errors.As(err, &anError)
	)

	return anError, ok
}
