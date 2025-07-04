package fsx

//spellchecker:words path filepath
import (
	"os"
	"path/filepath"
)

// Same checks if path1 and path2 refer to the same path.
// If both paths exist, they are compared using [os.Same].
// If both files do not exist, the paths are first compared syntactically and then via recursion on [filepath.Dir].
func Same(path1, path2 string) bool {
	// if the paths are identical, then we don't need to check anything.
	// and in particular, we don't need to do any expensive stat calls.
	if filepath.Clean(path1) == filepath.Clean(path2) {
		return true
	}

	// initial attempt: check if directly
	same, certain := couldBeSameFile(path1, path2)
	if certain {
		return same
	}

	// second attempt: find the directory names and base paths
	d1, n1 := filepath.Split(path1)
	d2, n2 := filepath.Split(path2)

	// if we have different file names we don't need to continue
	if n1 != n2 {
		return false
	}

	// compare the base names!
	{
		same, _ := couldBeSameFile(d1, d2)
		return same
	}
}

// couldBeSameFile checks if path1 might be the same as path2.
//
// If both files exist, compares using [os.SameFile].
// Otherwise compares absolute paths using string comparison.
//
// same indicates if they might be the same file.
// authoritative indicates if the result is authoritative.
func couldBeSameFile(path1, path2 string) (same, authoritative bool) {
	{
		info1, notExists1, err1 := stat(path1, true)
		info2, notExists2, err2 := stat(path2, true)

		// both files exist => check using os.SameFile
		// the result is always authoritative
		if err1 == nil && err2 == nil {
			same = os.SameFile(info1, info2)
			authoritative = true
			return
		}

		// only 1 file errored => they could be different
		if (err1 == nil) != (err2 == nil) {
			return
		}

		// only 1 file does not exist => they could be different
		if notExists1 != notExists2 {
			return
		}
	}

	{
		// resolve paths absolutely
		rpath1, err1 := filepath.Abs(path1)
		rpath2, err2 := filepath.Abs(path2)

		// if either path could not be resolved absolutely
		// fallback to just using clean!
		if err1 != nil {
			rpath1 = filepath.Clean(path1)
		}
		if err2 != nil {
			rpath2 = filepath.Clean(path2)
		}

		// compare using strings
		same = rpath1 == rpath2
		authoritative = same // positive result is authoritative!
		return
	}
}
