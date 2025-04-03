//spellchecker:words umaskfree
package umaskfree

//spellchecker:words errors time
import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"
)

//spellchecker:words nolint wrapcheck

// Create is like [os.Create] with an additional mode argument.
func Create(path string, mode fs.FileMode) (*os.File, error) {
	m.Lock()
	defer m.Unlock()

	//nolint:wrapcheck
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode) // #nosec G304 -- path is an explicit parameter
}

// WriteFile is like [os.WriteFile].
func WriteFile(path string, data []byte, mode fs.FileMode) (err error) {
	var handle *os.File
	handle, err = Create(path, mode)
	if err != nil {
		return err
	}

	defer func() {
		errClose := handle.Close()
		if err == nil {
			err = errClose
		}
	}()

	if _, err := handle.Write(data); err != nil {
		return err //nolint:wrapcheck
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
			return fmt.Errorf("failed to create file: %w", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
		return nil
	case err != nil:
		return fmt.Errorf("failed to stat path: %w", err)
	default:
		now := time.Now().Local()
		if err := os.Chtimes(path, now, now); err != nil {
			return fmt.Errorf("failed to change file time: %w", err)
		}
		return nil
	}
}

//spellchecker:words nosec
