//spellchecker:words password
package password

//spellchecker:words bufio embed strings sync
import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
)

//spellchecker:words fsys

// PasswordSource represents a source of passwords.
type PasswordSource interface {
	// Name returns the name of this source
	Name() string

	// Passwords returns a channel that reads all passwords.
	// If an error occurs, returns an empty channel
	//
	// The caller must drain the channel for it to be garbage collected.
	Passwords() <-chan string
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

func (cpr *commonPasswordReader) Passwords() <-chan string {
	src := make(chan string, 10)
	go func() {
		defer close(src)

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
			src <- line
		}
	}()
	return src
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

// Passwords returns a channel that contains all passwords in the provided sources.
// Passwords may be returned in any order.
//
// The caller must drain the channel.
func Passwords(sources ...PasswordSource) <-chan CommonPassword {
	// TODO: make this an iter.Seq
	common := make(chan CommonPassword, 10*len(sources))

	var wg sync.WaitGroup
	wg.Add(len(sources))

	for _, source := range sources {
		go func(source PasswordSource) {
			defer wg.Done()

			name := source.Name()

			for password := range source.Passwords() {
				common <- CommonPassword{
					Source:   name,
					Password: password,
				}
			}
		}(source)
	}

	go func() {
		defer close(common)
		wg.Wait()
	}()

	return common
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
