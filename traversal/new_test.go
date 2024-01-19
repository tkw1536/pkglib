// Package iterator provides Generic Iterator and Generator Interfaces.
package traversal_test

import (
	"errors"
	"fmt"

	"github.com/tkw1536/pkglib/traversal"
)

func ExampleNew() {
	iterator := traversal.New(func(generator traversal.Generator[string]) {
		if !generator.Yield("hello") {
			return
		}
		if !generator.Yield("world") {
			return
		}
	})
	defer iterator.Close()

	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())

	// Output: hello
	// world
	// <nil>
}

func ExampleNew_close() {
	iterator := traversal.New(func(generator traversal.Generator[string]) {
		if !generator.Yield("hello") {
			return
		}

		if !generator.Yield("world") {
			return
		}

		panic("never reached")
	})
	defer iterator.Close()

	// request a single item, then close the iterator!
	for iterator.Next() {
		fmt.Println(iterator.Datum())
		iterator.Close()
	}

	fmt.Println(iterator.Err())

	// Output: hello
	// <nil>
}

func ExampleNew_error() {
	iterator := traversal.New(func(generator traversal.Generator[string]) {
		if !generator.Yield("hello") {
			return
		}

		generator.YieldError(errors.New("something went wrong"))
	})
	defer iterator.Close()

	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}

	fmt.Println(iterator.Err())

	// Output: hello
	// something went wrong
}

func ExampleSlice() {
	iterator := traversal.Slice([]int{1, 2, 3, 4, 5})
	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())
	// Output: 1
	// 2
	// 3
	// 4
	// 5
	// <nil>
}

func ExampleMap() {
	iterator :=
		traversal.Map(
			traversal.Slice([]int{1, 2, 3, 4, 5}),
			func(value int) bool {
				return value%2 == 0
			},
		)
	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())
	// Output: false
	// true
	// false
	// true
	// false
	// <nil>
}

func ExampleConnect() {
	source := traversal.Slice([]int{0, 1, 2})
	dest := traversal.Connect(source, func(element int, sender traversal.Generator[int]) (ok bool) {
		sender.Yield(2 * element)
		sender.Yield(2*element + 1)
		return true
	})

	for dest.Next() {
		fmt.Println(dest.Datum())
	}
	fmt.Println(dest.Err())
	// Output: 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// <nil>
}

func ExamplePipe() {
	iterator := traversal.New(func(generator traversal.Generator[int]) {
		source := traversal.Slice([]int{0, 1, 2, 3, 4, 5})
		traversal.Pipe(generator, source)
	})

	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())
	// Output: 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// <nil>
}

func ExampleDrain() {
	source := traversal.Slice([]int{0, 1, 2, 3, 4, 5})
	fmt.Println(traversal.Drain(source))
	// Output: [0 1 2 3 4 5] <nil>
}
