//spellchecker:words nobufio
package nobufio

//spellchecker:words golang term
import (
	"os"

	"golang.org/x/term"
)

type fdAble interface {
	Fd() uintptr
}

// GetTerminalFD returns the file descriptor of the terminal represented by the given stream.
// If stream does not represent a terminal, returns (0, false).
func GetTerminalFD(stream any) (fd int, ok bool) {
	file, ok := stream.(fdAble)
	if !ok {
		return 0, false
	}
	fd = int(file.Fd())
	if !term.IsTerminal(fd) {
		return 0, false
	}
	return fd, true
}

// IsTerminal checks if stream refers to a file descriptor that is a UNIX-like terminal.
// A refers to a file descriptor if it is of type *os.File.
func IsTerminal(stream any) bool {
	file, ok := stream.(*os.File)
	return ok && isTerminal(file)
}

func isTerminal(file *os.File) bool {
	return term.IsTerminal(int(file.Fd()))
}
