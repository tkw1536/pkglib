//spellchecker:words status
package status_test

//spellchecker:words testing github pkglib status
import (
	"fmt"
	"testing"

	"go.tkw01536.de/pkglib/status"
)

func ExampleLineBuffer() {
	// create a new line buffer
	buffer := status.LineBuffer{
		Line: func(line string) {
			fmt.Printf("Line(%q)\n", line)
		},
		CloseLine: func() {
			fmt.Println("CloseLine()")
		},
	}

	// write some text into it, calling Line() with each completed line
	_, _ = buffer.WriteString("line 1\npartial")
	_, _ = buffer.WriteString(" line 2\n\n line not terminated")

	// close the buffer, calling CloseLine()
	_ = buffer.Close()

	// futures writes are no longer calling Line
	_, _ = buffer.WriteString("another\nline\n")

	// Output: Line("line 1")
	// Line("partial line 2")
	// Line("")
	// CloseLine()
}

func ExampleLineBuffer_FlushLineOnClose() {
	// create a new line buffer
	buffer := status.LineBuffer{
		Line: func(line string) {
			fmt.Printf("Line(%q)\n", line)
		},
		FlushLineOnClose: true,

		CloseLine: func() {
			fmt.Println("CloseLine()")
		},
	}

	// write some text into it, calling Line() with each completed line
	_, _ = buffer.WriteString("line 1\npartial")
	_, _ = buffer.WriteString(" line 2\n\n line not terminated")

	// close the buffer, calling CloseLine()
	_ = buffer.Close()

	// futures writes are no longer calling Line
	_, _ = buffer.WriteString("another\nline\n")

	// Output: Line("line 1")
	// Line("partial line 2")
	// Line("")
	// Line(" line not terminated")
	// CloseLine()
}

func BenchmarkLineBuffer(b *testing.B) {
	buffer := status.LineBuffer{
		Line: func(line string) {
			/* do nothing */
		},
	}

	data := []byte("world\nhello")

	for b.Loop() {
		_, _ = buffer.Write(data)
	}
}
