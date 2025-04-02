//spellchecker:words traversal
package traversal_test

//spellchecker:words github pkglib traversal
import (
	"fmt"
	"iter"

	"github.com/tkw1536/pkglib/traversal"
)

func ExampleSequence() {
	seq := iter.Seq[int](func(yield func(int) bool) {
		if !yield(42) {
			return
		}
		if !yield(69) {
			return
		}
	})

	// turn it into an iterator and drain it!
	it := traversal.Sequence(seq)
	fmt.Println(traversal.Drain(it))

	// Output: [42 69] <nil>
}

func ExampleRange() {
	// create a simple iterator
	it := traversal.New(func(generator traversal.Generator[int]) {
		if !generator.Yield(42) {
			return
		}
		if !generator.Yield(69) {
			return
		}
	})

	rg := traversal.Range(it)
	rg(func(value int) bool {
		fmt.Println(value)
		return true
	})

	// Output: 42
	// 69
}
