//spellchecker:words nobufio
package nobufio_test

//spellchecker:words strings testing unicode pkglib nobufio
import (
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"go.tkw01536.de/pkglib/nobufio"
)

// runes of different sizes to test ReadRune() for.
var runes = []rune{
	'a', 'A', '1', '!', ' ',
	'é', 'ø', 'ç',
	'€', 'あ', '中',
	'𝔘', '😀',
}

func TestReadRune_NoReadRuneAvailable(t *testing.T) {
	t.Parallel()

	for _, char := range runes {
		t.Run(string(char), func(t *testing.T) {
			t.Parallel()

			expectedSize := utf8.RuneLen(char)

			reader := &MaxBytesReader{
				Reader: strings.NewReader(string(char)),
				Bytes:  expectedSize,
			}
			r, size, err := nobufio.ReadRune(reader)
			if err != nil {
				t.Fatalf("ReadRune failed: %v", err)
			}

			if r != char {
				t.Errorf("Expected rune %q, got %q", char, r)
			}

			if size != expectedSize {
				t.Errorf("Expected size %d, got %d", expectedSize, size)
			}

			if reader.Bytes != 0 {
				t.Errorf("Expected %d bytes left, got %d", 0, reader.Bytes)
			}
		})
	}
}

func TestReadRune_RuneReaderAvailable(t *testing.T) {
	t.Parallel()

	for _, char := range runes {
		t.Run(string(char), func(t *testing.T) {
			t.Parallel()

			expectedSize := utf8.RuneLen(char)

			reader := &OnlyRuneReader{
				Reader: strings.NewReader(string(char) + "more text here"),
			}
			r, size, err := nobufio.ReadRune(reader)
			if err != nil {
				t.Fatalf("ReadRune failed: %v", err)
			}

			if r != char {
				t.Errorf("Expected rune %q, got %q", char, r)
			}

			if size != expectedSize {
				t.Errorf("Expected size %d, got %d", expectedSize, size)
			}
		})
	}
}

// MaxBytesReader is a reader that allows to read up to a specific number of bytes.
// Calling Read() with a too large buffer panics().
type MaxBytesReader struct {
	Reader io.Reader
	Bytes  int
}

func (m *MaxBytesReader) Read(p []byte) (int, error) {
	if len(p) > m.Bytes {
		panic("MaxBytesReader.Read(): Not allowed")
	}
	m.Bytes -= len(p)
	bytes, err := m.Reader.Read(p)
	if err != nil {
		return bytes, fmt.Errorf("failed to read from reader: %w", err)
	}
	return bytes, nil
}

// OnlyRuneReader is a reader that allows calls to ReadRune(), but not Read().
type OnlyRuneReader struct {
	Reader io.RuneReader
}

func (o *OnlyRuneReader) ReadRune() (rune, int, error) {
	r, size, err := o.Reader.ReadRune()
	if err != nil {
		return r, size, fmt.Errorf("failed to read rune from reader: %w", err)
	}
	return r, size, nil
}

func (o *OnlyRuneReader) Read(p []byte) (int, error) {
	panic("OnlyRuneReader.Read(): Not allowed")
}
