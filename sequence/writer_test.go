// Package sequence provides Writer.
package sequence

import (
	"errors"
	"fmt"
	"os"
)

func ExampleWriter() {
	// Writer counts the total number of bytes written
	w := Writer{
		Writer: os.Stdout,
	}

	w.Write([]byte("hello world\n"))
	w.Write([]byte("bye world\n"))

	// the total number of bytes written and any error
	fmt.Println(w.Sum())

	// Output: hello world
	// bye world
	// 22 <nil>
}

// finitewrites is a writer that writes to stdout a finite number of times.
type finitewrites int

func (f *finitewrites) Write(d []byte) (int, error) {
	if f == nil || *f <= 0 {
		return 0, errors.New("no writes left")
	}
	*f--
	return os.Stdout.Write(d)
}

func ExampleWriter_fail() {
	// create a writer that can only be written to once
	writeonce := finitewrites(1)

	w := Writer{
		Writer: &writeonce,
	}

	// write to it twice
	n1, err1 := w.Write([]byte("hello world\n")) // write will work
	n2, err2 := w.Write([]byte("bye world\n"))   // write will fail
	n3, err3 := w.Write([]byte("hello mars\n"))  // write will fail

	fmt.Println(n1, err1)
	fmt.Println(n2, err2)
	fmt.Println(n3, err3)
	fmt.Println(w.Sum())

	// Output: hello world
	// 12 <nil>
	// 0 no writes left
	// 0 previous write failed: no writes left
	// 12 no writes left
}
