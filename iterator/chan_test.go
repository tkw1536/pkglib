package iterator

import (
	"context"
	"fmt"
)

func ExampleAsChannel() {
	it := Slice([]int{0, 1, 2, 3, 4, 5})

	for i := range AsChannel(it, context.Background()) {
		fmt.Print(i, " ")
	}

	// Output: 0 1 2 3 4 5
}

func ExampleFromChannel() {
	// fill a channel with some numbers
	// and then close it!
	in := make(chan int, 6)
	for i := 0; i < 6; i++ {
		in <- i
	}
	close(in)

	// create a channel
	c, _ := FromChannel(in)

	for c.Next() {
		fmt.Print(c.Datum(), " ")
	}

	// Output: 0 1 2 3 4 5
}
