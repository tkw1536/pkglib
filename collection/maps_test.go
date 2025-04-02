//spellchecker:words collection
package collection

import (
	"fmt"
)

func ExampleIterSorted() {
	m := map[int]string{
		0: "hello",
		1: "world",
	}

	for k, v := range IterSorted(m) {
		fmt.Printf("%d: %v\n", k, v)
	}

	// Output: 0: hello
	// 1: world
}

func ExampleIterateSorted() {
	m := map[int]string{
		0: "hello",
		1: "world",
		2: "i",
		3: "like",
		4: "you",
	}

	IterateSorted(m, func(k int, v string) bool {
		fmt.Printf("%d: %v\n", k, v)

		// We return false to stop the iteration.
		// Here this happens at k = 1.
		return k < 1
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

func ExampleAppend() {
	// append to the first map
	fmt.Println(Append(map[string]string{
		"hello": "world",
	}, map[string]string{
		"answer": "42",
	}))

	// append to the first non-nil map
	fmt.Println(Append(nil, nil, nil, map[string]string{
		"hello": "world",
	}, map[string]string{
		"answer": "42",
	}))

	// appending nothing results in an empty map
	fmt.Println(Append[string, string]())

	// appending to the nil map results in an empty map
	fmt.Println(Append[string, string](nil))

	// Output: map[answer:42 hello:world]
	// map[answer:42 hello:world]
	// map[]
	// map[]
}
