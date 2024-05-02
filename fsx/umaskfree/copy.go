//spellchecker:words umaskfree
package umaskfree

//spellchecker:words context errors path filepath github pkglib contextx
import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tkw1536/pkglib/contextx"
	"github.com/tkw1536/pkglib/fsx"
)

var ErrCopySameFile = errors.New("src and dst must be different")

// CopyFile copies a file from src to dst.
// When src points to a symbolic link, will copy the symbolic link.
//
// When dst and src are the same file, returns [ErrCopySameFile].
// When ctx is closed, the file is not copied.
func CopyFile(ctx context.Context, dst, src string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if fsx.Same(src, dst) {
		return ErrCopySameFile
	}

	// open the source
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// stat it to get the mode!
	srcStat, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// open or create the destination
	dstFile, err := Create(dst, srcStat.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// and do the copy!
	_, err = contextx.Copy(ctx, dstFile, srcFile)
	return err
}

// CopyLink copies a link from src to dst.
// If dst already exists, it is deleted and then re-created.
func CopyLink(ctx context.Context, dst, src string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// if they're the same file that is an error
	if fsx.Same(dst, src) {
		return ErrCopySameFile
	}

	// read the link target
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}

	// delete it if it already exists
	{
		exists, err := fsx.Exists(dst)
		if err != nil {
			return err
		}
		if exists {
			if err := os.Remove(dst); err != nil {
				return err
			}
		}
	}

	// make the symbolic link!
	return os.Symlink(target, dst)
}

var ErrDstFile = errors.New("dst is a file")

// CopyDirectory copies the directory src to dst recursively.
// Copying is aborted when ctx is closed.
//
// Existing files and directories are overwritten.
// When a directory already exists, additional files are not deleted.
//
// onCopy, when not nil, is called for each file or directory being copied.
func CopyDirectory(ctx context.Context, dst, src string, onCopy func(dst, src string)) error {
	// sanity checks
	if fsx.Same(src, dst) {
		return ErrCopySameFile
	}

	// check that the destination is a regular file
	{
		isRegular, err := fsx.IsRegular(dst, true)
		if err != nil {
			return err
		}
		if isRegular {
			return ErrDstFile
		}
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		// someone previously returned an error
		if err != nil {
			return err
		}

		// context was closed
		if err := ctx.Err(); err != nil {
			return err
		}

		// determine the real target path
		var relpath string
		relpath, err = filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dst, relpath)

		// call the hook
		if onCopy != nil {
			onCopy(dst, src)
		}

		// stat the directory, so that we can get mode, and info later!
		info, err := d.Info()
		if err != nil {
			return err
		}

		// if we have a symbolic link, copy the link!
		if info.Mode()&fs.ModeSymlink != 0 {
			return CopyLink(ctx, dst, path)
		}

		// if we got a file, we should copy it normally
		if !d.IsDir() {
			return CopyFile(ctx, dst, path)
		}

		// create the directory, but ignore an error if the directory already exists.
		// this is so that we can copy one tree into another tree.
		err = Mkdir(dst, info.Mode())
		if errors.Is(err, fs.ErrExist) {
			isDir, isDirE := fsx.IsDirectory(dst, false)
			if isDirE != nil {
				return isDirE
			}
			if isDir {
				err = nil
			}
		}

		return err
	})
}
