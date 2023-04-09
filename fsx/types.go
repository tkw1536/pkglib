package fsx

import (
	"errors"
	"io/fs"
	"os"
)

// stat performs stat on path, following links if requested.
func stat(path string, follow bool) (info fs.FileInfo, IsNotNotExists bool, err error) {
	if follow {
		info, err = os.Stat(path)
	} else {
		info, err = os.Lstat(path)
	}
	IsNotNotExists = err != nil && !errors.Is(err, fs.ErrNotExist)
	return
}

// Exists checks if the given path exists.
// An invalid link is considered to exist.
//
// If an error occurs, returns false, err.
func Exists(path string) (bool, error) {
	_, other, err := stat(path, false)
	if other {
		return false, err
	}
	return err == nil, nil
}

// IsDirectory checks if the provided path exists and is a directory.
// IsDirectory follows links iff followLinks is true.
func IsDirectory(path string, followLinks bool) (bool, error) {
	info, other, err := stat(path, followLinks)
	if other {
		return false, err
	}
	return err == nil && info.Mode().IsDir(), nil
}

// IsRegular checks if the provided path exists and is a directory.
// IsRegular follows links iff followLinks is true.
func IsRegular(path string, followLinks bool) (bool, error) {
	info, other, err := stat(path, followLinks)
	if other {
		return false, err
	}
	return err == nil && info.Mode().IsRegular(), nil
}

// IsLink checks if the provided path exists and is a symlink.
// An invalid link is considered a link.
func IsLink(path string) (bool, error) {
	info, other, err := stat(path, false)
	if other {
		return false, err
	}
	return err == nil && info.Mode()&fs.ModeSymlink != 0, nil
}
