// Package sequence provides Writer.
//
//spellchecker:words sequence
package sequence_test

//spellchecker:words errors sync github pkglib sequence
import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/tkw1536/pkglib/sequence"
)

func ExampleWriter() {
	// Writer counts the total number of bytes written
	w := sequence.Writer{
		Writer: os.Stdout,
	}

	_, _ = w.Write([]byte("hello world\n"))
	_, _ = w.Write([]byte("bye world\n"))

	// the total number of bytes written and any error
	fmt.Println(w.Sum())

	// Output: hello world
	// bye world
	// 22 <nil>
}

var (
	errNoWritesLeft    = errors.New("no writes left")
	errConcurrentWrite = errors.New("concurrent write not allowed")
)

// writeSyncToStdout is a writer that allows only a finite number of writes.
// Concurrent writes are an error.
// Writes are passed to stdout.
type writeSyncToStdout struct {
	NumWrites int
	l         sync.Mutex
}

func (f *writeSyncToStdout) Write(d []byte) (int, error) {
	// lock and bail out if concurrent
	if !f.l.TryLock() {
		return 0, errConcurrentWrite
	}
	defer f.l.Unlock()

	// check that we have some amount of Write() calls left
	if f.NumWrites <= 0 {
		return 0, errNoWritesLeft
	}

	// do the write!
	f.NumWrites--

	n, err := os.Stdout.Write(d)
	if err != nil {
		return n, fmt.Errorf("failed to write to stdout: %w", err)
	}
	return n, nil
}

func ExampleWriter_concurrent() {
	// create a writer that can only be written to once
	writeOnce := writeSyncToStdout{NumWrites: 2}

	w := sequence.Writer{
		Writer: &writeOnce,
	}

	// write a bunch of times concurrently
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Write([]byte("something\n"))
		}()
	}

	wg.Wait()
	fmt.Println(w.Sum())

	// Output: something
	// something
	// 20 no writes left
}

func ExampleWriter_fail() {
	// create a writer that can only be written to once
	writeOnce := writeSyncToStdout{NumWrites: 1}

	w := sequence.Writer{
		Writer: &writeOnce,
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
