//spellchecker:words umaskfree
package umaskfree

//spellchecker:words errors time
import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// Create is like [os.Create] with an additional mode argument.
func Create(path string, mode fs.FileMode) (*os.File, error) {
	m.Lock()
	defer m.Unlock()

	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode) // #nosec G304 -- path is an explicit parameter
}

// WriteFile is like [os.WriteFile].
func WriteFile(path string, data []byte, mode fs.FileMode) error {
	handle, err := Create(path, mode)
	if err != nil {
		return err
	}
	defer handle.Close()

	if _, err := handle.Write(data); err != nil {
		return err
	}

	return nil
}

// DefaultFilePerm should be used by callers to use a consistent file mode for new files.
const DefaultFilePerm fs.FileMode = 0666

// Touch touches a file.
// It is similar to the unix 'touch' command.
//
// If the file does not exist, it is created using [Create].
// If the file does exist, its' access and modification times are updated to the current time.
func Touch(path string, perm fs.FileMode) error {
	if perm == 0 {
		perm = DefaultFilePerm
	}
	_, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		f, err := Create(path, perm)
		if err != nil {
			return err
		}
		defer f.Close()
		return nil
	case err != nil:
		return err
	default:
		now := time.Now().Local()
		return os.Chtimes(path, now, now)
	}
}

// spellchecker:words nosec
