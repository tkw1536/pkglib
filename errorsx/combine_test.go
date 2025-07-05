//spellchecker:words errorsx
package errorsx_test

//spellchecker:words errors reflect testing github pkglib errorsx
import (
	"errors"
	"reflect"
	"testing"

	"go.tkw01536.de/pkglib/errorsx"
)

// This test code is heavily adapted from [errors.Join].
var (
	err1 = errors.New("err1")
	err2 = errors.New("err2")
)

func TestCombineReturnsNil(t *testing.T) {
	t.Parallel()

	if err := errorsx.Combine(); err != nil {
		t.Errorf("errorsx.Combine() = %v, want nil", err)
	}
	if err := errorsx.Combine(nil); err != nil {
		t.Errorf("errorsx.Combine(nil) = %v, want nil", err)
	}
	if err := errorsx.Combine(nil, nil); err != nil {
		t.Errorf("errorsx.Combine(nil, nil) = %v, want nil", err)
	}
}

func TestCombine(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		errs []error
		want []error
	}{{
		errs: []error{err1},
		want: nil,
	}, {
		errs: []error{err1, err2},
		want: []error{err1, err2},
	}, {
		errs: []error{err1, nil, err2},
		want: []error{err1, err2},
	}, {
		errs: []error{err1, nil, nil},
		want: nil,
	}} {
		unwrap, ok := errorsx.Combine(test.errs...).(interface{ Unwrap() []error })
		var got []error
		if ok {
			got = unwrap.Unwrap()
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Combine(%v) = %v; want %v", test.errs, got, test.want)
		}
		if len(got) != cap(got) {
			t.Errorf("Combine(%v) returns errors with len=%v, cap=%v; want len==cap", test.errs, len(got), cap(got))
		}
	}
}

func TestCombineErrorMethod(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		errs []error
		want string
	}{{
		errs: []error{err1},
		want: "err1",
	}, {
		errs: []error{err1, nil, nil},
		want: "err1",
	}, {
		errs: []error{err1, err2},
		want: "err1\nerr2",
	}, {
		errs: []error{err1, nil, err2},
		want: "err1\nerr2",
	}} {
		got := errorsx.Combine(test.errs...).Error()
		if got != test.want {
			t.Errorf("Combine(%v).Error() = %q; want %q", test.errs, got, test.want)
		}
	}
}
