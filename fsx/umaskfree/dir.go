//spellchecker:words umaskfree
package umaskfree

//spellchecker:words path filepath syscall
import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

//spellchecker:words nolint wrapcheck

// DefaultDirPerm should be used by callers to use a consistent mode for new directories.
const DefaultDirPerm fs.FileMode = fs.ModeDir | fs.ModePerm

// Mkdir is like [os.Mkdir].
func Mkdir(path string, perm fs.FileMode) error {
	targetPerm := fs.ModeDir | perm

	if err := os.Mkdir(path, targetPerm); err != nil {
		return fmt.Errorf("failed to make directory: %w", err)
	}
	if err := os.Chmod(path, targetPerm); err != nil {
		return fmt.Errorf("failed to set mode: %w", err)
	}
	return nil
}

// MkdirAll is like [os.MkdirAll].
func MkdirAll(path string, perm fs.FileMode) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := os.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return fmt.Errorf("failed to create directory: %w", &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR})
	}

	// Slow path: make sure parent exists and then call Mkdir for path.

	// Extract the parent folder from path by first removing any trailing
	// path separator and then scanning backward until finding a path
	// separator or reaching the beginning of the string.
	i := len(path) - 1
	for i >= 0 && os.IsPathSeparator(path[i]) {
		i--
	}
	for i >= 0 && !os.IsPathSeparator(path[i]) {
		i--
	}
	if i < 0 {
		i = 0
	}

	// If there is a parent directory, and it is not the volume name,
	// recurse to ensure parent directory exists.
	if parent := path[:i]; len(parent) > len(filepath.VolumeName(path)) {
		err = MkdirAll(parent, perm)
		if err != nil {
			return fmt.Errorf("failed to make directories: %w", err)
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := os.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	return nil
}
