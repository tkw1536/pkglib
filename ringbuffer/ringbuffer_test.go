package ringbuffer_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/ringbuffer"
)

// Demonstrates how a ring buffer works by adding to it
func ExampleRingBuffer() {
	buffer := ringbuffer.MakeRingBuffer[string](2)

	buffer.Add("hello")
	buffer.Add("world")
	fmt.Println(buffer.Elems())

	buffer.Add("how")
	fmt.Println(buffer.Elems())

	buffer.Add("are")
	fmt.Println(buffer.Elems())

	buffer.Add("you")
	fmt.Println(buffer.Elems())

	// Output:
	// [hello world]
	// [world how]
	// [how are]
	// [are you]
}

func ExampleRingBuffer_Push() {
	buffer := ringbuffer.MakeRingBuffer[string](0)

	// adding to a zero capacity buffer is a no-op
	buffer.Add("to be ignored")
	fmt.Println(buffer.Elems())

	// push a new word always adds to the end of the buffer
	buffer.Push("hello")
	fmt.Println(buffer.Elems())

	// Output: []
	// [hello]
}

func TestRingBuffer_elems(t *testing.T) {
	for i := range 100 {
		tt := struct{ iterations, bufsize int }{
			iterations: 1000,
			bufsize:    i,
		}
		t.Run(fmt.Sprintf("iterations: %d bufsize: %d", tt.iterations, tt.bufsize), func(t *testing.T) {
			buffer := ringbuffer.MakeRingBuffer[int](tt.bufsize)
			wantElems := make([]int, tt.bufsize)
			for i := range tt.iterations {
				buffer.Add(i)
				if i < tt.bufsize {
					continue
				}

				// setup what we want from the elements
				for e := range wantElems {
					wantElems[e] = i - (tt.bufsize - 1 - e)
				}

				gotElems := buffer.Elems()
				if !reflect.DeepEqual(gotElems, wantElems) {
					t.Errorf("got elems = %v, want elems = %v", gotElems, wantElems)
				}
			}
		})
	}
}
