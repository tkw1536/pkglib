//spellchecker:words traversal
package traversal_test

//spellchecker:words context github pkglib traversal
import (
	"context"
	"fmt"

	"github.com/tkw1536/pkglib/traversal"
)

func ExampleAsChannel() {
	it := traversal.Slice([]int{0, 1, 2, 3, 4, 5})

	for i := range traversal.AsChannel(it, context.Background()) {
		fmt.Print(i, " ")
	}

	// Output: 0 1 2 3 4 5
}

func ExampleFromChannel() {
	// fill a channel with some numbers
	// and then close it!
	in := make(chan int, 6)
	for i := range cap(in) {
		in <- i
	}
	close(in)

	// create a channel
	c, _ := traversal.FromChannel(in)

	for c.Next() {
		fmt.Print(c.Datum(), " ")
	}

	// Output: 0 1 2 3 4 5
}
