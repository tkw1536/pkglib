//spellchecker:words password
package password

//spellchecker:words bufio embed iter strings
import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"strings"
)

//spellchecker:words fsys

// PasswordSource represents a source of passwords.
type PasswordSource interface {
	// Name returns the name of this source
	Name() string

	// Passwords returns a list of passwords.
	Passwords() iter.Seq[string]
}

// NewPasswordSource creates a new password source from a function.
func NewPasswordSource(open func() (io.Reader, error), name string) PasswordSource {
	return &commonPasswordReader{
		open: open,
		name: name,
	}
}

// NewSources reads a set of sources from a file system.
// All files matching pattern are returned.
func NewSources(fsys fs.FS, pattern string) ([]PasswordSource, error) {
	matches, err := fs.Glob(fsys, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to apply pattern: %w", err)
	}

	sources := make([]PasswordSource, len(matches))
	for i, match := range matches {
		sources[i] = NewPasswordSource(func() (io.Reader, error) { return fsys.Open(match) }, match)
	}

	return sources, nil
}

type commonPasswordReader struct {
	open func() (io.Reader, error)
	name string
}

func (cpr *commonPasswordReader) Name() string {
	return cpr.name
}

func (cpr *commonPasswordReader) Passwords() iter.Seq[string] {
	return func(yield func(string) bool) {
		file, err := cpr.open()
		if err != nil {
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			if !yield(line) {
				return
			}
		}
	}
}

type CommonPassword struct {
	Password string
	Source   string
}

// CommonPasswordError.
type CommonPasswordError struct {
	CommonPassword
}

func (cpe CommonPasswordError) Error() string {
	return fmt.Sprintf("%q from %q", cpe.Password, cpe.Source)
}

//go:embed common
var commonEmbed embed.FS

// CommonSources returns a list of common password sources.
func CommonSources() []PasswordSource {
	sources, err := NewSources(commonEmbed, "**/*.txt")
	if err != nil {
		panic(err)
	}
	return sources
}

// Passwords returns an iterator that iterates over all of the passwords in all of the sources.
// They may be returned in any order.
func Passwords(sources ...PasswordSource) iter.Seq[CommonPassword] {
	return func(yield func(CommonPassword) bool) {
		for _, source := range sources {
			name := source.Name()

			for password := range source.Passwords() {
				if !yield(CommonPassword{
					Source:   name,
					Password: password,
				}) {
					return
				}
			}
		}
	}
}

// - or nil (when a password is not a common password.
func CheckCommonPassword(check func(candidate string) (bool, error), sources ...PasswordSource) error {
	for common := range Passwords(sources...) {
		ok, err := check(common.Password)
		if err != nil {
			return err
		}

		// password validation passed
		if ok {
			return CommonPasswordError{
				CommonPassword: common,
			}
		}
	}
	return nil
}
