package status

import (
	"fmt"
	"testing"
)

func ExampleLineBuffer() {
	// create a new line buffer
	buffer := LineBuffer{
		Line: func(line string) {
			fmt.Printf("Line(%q)\n", line)
		},
		CloseLine: func() {
			fmt.Println("CloseLine()")
		},
	}

	// write some text into it, calling Line() with each completed line
	buffer.WriteString("line 1\npartial")
	buffer.WriteString(" line 2\n\n line not terminated")

	// close the buffer, calling CloseLine()
	buffer.Close()

	// futures writes are no longer calling Line
	buffer.WriteString("another\nline\n")

	// Output: Line("line 1")
	// Line("partial line 2")
	// Line("")
	// CloseLine()
}

func ExampleLineBuffer_FlushLineOnClose() {
	// create a new line buffer
	buffer := LineBuffer{
		Line: func(line string) {
			fmt.Printf("Line(%q)\n", line)
		},
		FlushLineOnClose: true,

		CloseLine: func() {
			fmt.Println("CloseLine()")
		},
	}

	// write some text into it, calling Line() with each completed line
	buffer.WriteString("line 1\npartial")
	buffer.WriteString(" line 2\n\n line not terminated")

	// close the buffer, calling CloseLine()
	buffer.Close()

	// futures writes are no longer calling Line
	buffer.WriteString("another\nline\n")

	// Output: Line("line 1")
	// Line("partial line 2")
	// Line("")
	// Line(" line not terminated")
	// CloseLine()
}

func BenchmarkLineBuffer(b *testing.B) {

	buffer := LineBuffer{
		Line: func(line string) {
			/* do nothing */
		},
	}

	data := []byte("world\nhello")

	for i := 0; i < b.N; i++ {
		buffer.Write(data)
	}
}
