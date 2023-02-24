package collection

import (
	"fmt"
)

func ExampleIterateSorted() {
	m := map[int]string{
		0: "hello",
		1: "world",
	}
	IterateSorted(m, func(k int, v string) {
		fmt.Printf("%d: %v\n", k, v)
	})

	// Output: 0: hello
	// 1: world
}

func ExampleMapValues() {
	m := map[int]string{
		0: "hi",
		1: "world",
	}
	m2 := MapValues(m, func(k int, v string) int {
		return len(v)
	})
	fmt.Println(m2)

	// Output: map[0:2 1:5]
}
