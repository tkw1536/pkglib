package testlib

import (
	"path/filepath"
	"testing"
)

// TempDirAbs is like the TempDir method of t, but resolves all symlinks in the returned path.
func TempDirAbs(t *testing.T) string {
	// NOTE(twiesing): This function is untested
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		panic(err)
	}
	return path
}
