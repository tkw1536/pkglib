package pkglib

import (
	"runtime"
	"strings"
	"testing"
)

func TestFutureTODOs(t *testing.T) {
	t.Parallel()

	if strings.Trim(runtime.Version(), "go") > "1.25.0" {
		t.Error("must migrate to synctest")
		// TODO: remove the build tags on existing files
		// TODO: migrate everything using time.Sleep in tests there
	}
}
