package testlib_test

import (
	"strconv"
	"testing"

	"go.tkw01536.de/pkglib/testlib"
)

func TestProduceExitError(t *testing.T) {
	t.Parallel()

	for i := range 128 {
		if i == 0 {
			continue
		}
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ee := testlib.ProduceExitError(i)
			if ee.ExitCode() != i {
				t.Errorf("expected exit code %d; got %d", i, ee.ExitCode())
			}
		})
	}
}
