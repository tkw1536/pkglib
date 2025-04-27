//spellchecker:words errorsx
package errorsx

import (
	"fmt"
	"io"
)

//spellchecker:words retval

// Close closes the given closer and updates retval if closing failed.
// if retval is nil, Close panic()s.
// desc is used as the description of closer.
//
// Close is intended to be deferred:
//
//	func stuff() (e error) {
//		f, err := os.Open(...)
//		if err != nil { /* ... */ }
//		defer errorsx.Close(f, &e, "file")
//		/* ... */
//	}
func Close(closer io.Closer, retval *error, desc string) {
	if retval == nil {
		panic("Close: nil retval should be replaced by a plain .Close() call")
	}

	err := closer.Close()
	if err == nil {
		return
	}

	*retval = Combine(*retval, &closeError{name: desc, err: err})
}

type closeError struct {
	name string
	err  error
}

func (ftc *closeError) Error() string {
	return fmt.Sprintf("failed to close %s: %s", ftc.name, ftc.err)
}

func (ftc *closeError) Unwrap() error {
	return ftc.err
}
