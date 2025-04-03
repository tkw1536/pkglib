//spellchecker:words ringbuffer
package ringbuffer_test

//spellchecker:words reflect testing github pkglib ringbuffer
import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/ringbuffer"
)

// Demonstrates how a ring buffer works by adding to it.
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

func TestRingBuffer(t *testing.T) {
	t.Parallel()

	// this function tests everything for the RingBuffer
	// with the exception of the Push function.

	BufSizeLimit := 10
	IterationCount := 1000

	for i := range BufSizeLimit {
		for _, optimized := range []bool{false, true} {
			tt := struct {
				iterations, bufferSize int
				optimized              bool
			}{
				iterations: IterationCount,
				bufferSize: i,
				optimized:  optimized,
			}
			t.Run(fmt.Sprintf("iterations: %d bufferSize: %d optimized %t", tt.iterations, tt.bufferSize, tt.optimized), func(t *testing.T) {
				t.Parallel()

				primary := ringbuffer.MakeRingBuffer[int](tt.bufferSize)
				for iterationNo := range tt.iterations {

					// compute the state we want
					wantElems := make([]int, 0, tt.bufferSize)
					if iterationNo >= tt.bufferSize {
						for e := range tt.bufferSize {
							// ringbuffer is full
							// so it should contain the most recent tt.bufferSize elements
							wantElems = append(wantElems, iterationNo-(tt.bufferSize-1-e))
						}
					} else {
						// ringbuffer is not yet full
						// so it should contain all elements
						for i := range iterationNo + 1 {
							wantElems = append(wantElems, i)
						}
					}

					t.Run(fmt.Sprintf("primary buffer iteration %d", iterationNo), func(t *testing.T) {
						primary.Add(iterationNo)
						if tt.optimized {
							primary.Optimize()
						}
						checkBufferState(t, BufferState[int]{Elems: wantElems, Cap: tt.bufferSize}, primary)
					})

					t.Run(fmt.Sprintf("pop buffer iteration %d", iterationNo), func(t *testing.T) {
						popBuffer := ringbuffer.MakeRingBuffer[int](tt.bufferSize)
						for i := range iterationNo + 1 {
							popBuffer.Add(i)
						}

						// pop all the elements, one by one
						for i := len(wantElems) - 1; i >= 0; i-- {
							got := popBuffer.Pop()
							if wantElems[i] != got {
								t.Errorf("pop got element = %v, want element = %v", wantElems[i], got)
							}

							if tt.optimized {
								popBuffer.Optimize()
							}

							checkBufferState(t, BufferState[int]{Elems: wantElems[:i], Cap: tt.bufferSize}, popBuffer)
						}

					})
				}
			})
		}
	}
}

type BufferState[T any] struct {
	Elems  []T
	Cap    int
	CapMin bool
}

func checkBufferState[T any](tb testing.TB, want BufferState[T], buffer *ringbuffer.RingBuffer[T]) {
	tb.Helper()

	wantLen := len(want.Elems)
	wantCap := want.Cap

	gotLen := buffer.Len()
	if gotLen != wantLen {
		tb.Errorf("got len = %v, want len = %v", gotLen, wantLen)
	}

	gotCap := buffer.Cap()
	if !want.CapMin {
		if gotCap != wantCap {
			tb.Errorf("got cap = %v, want cap = %v", gotCap, wantCap)
		}
	} else {
		if gotCap < wantCap {
			tb.Errorf("got cap = %v, want cap >= %v", gotCap, wantCap)
		}
	}

	gotElems := buffer.Elems()
	if !reflect.DeepEqual(gotElems, want.Elems) {
		tb.Errorf("got elems = %v, want elems = %v", gotElems, want.Elems)
	}

	var nextIndex int
	buffer.Iterate(func(elem T, index int) bool {
		if index != nextIndex {
			tb.Errorf("Iterate called in the wrong order, expected %d but got %d", nextIndex, index)
		}
		nextIndex++

		if index < 0 || index >= len(want.Elems) {
			tb.Errorf("Iterate called with unexpected element %v", elem)
			return true
		}

		wantElem := want.Elems[index]
		if !reflect.DeepEqual(elem, wantElem) {
			tb.Errorf("Iterate called with wrong element, expected %v but got %v", wantElem, elem)
		}

		return true
	})
	if nextIndex != len(gotElems) {
		tb.Errorf("Iterate called an unexpected number of times, expected %d but got %d", len(gotElems), nextIndex)
	}

}

func TestRingBuffer_Push(t *testing.T) {
	t.Parallel()

	N := 1000
	buffer := ringbuffer.MakeRingBuffer[int](0)

	checkBufferState(t, BufferState[int]{
		Elems: []int{},
		Cap:   0,
	}, buffer)

	for i := range N {
		buffer.Push(i)

		want := make([]int, 0, i+1)
		for j := range i + 1 {
			want = append(want, j)
		}

		checkBufferState(t, BufferState[int]{
			Elems:  want,
			Cap:    i + 1,
			CapMin: true,
		}, buffer)
	}

}
