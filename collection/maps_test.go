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

	fmt.Println(Append[string, string]())

	// Output: map[answer:42 hello:world]
	// map[answer:42 hello:world]
	// map[]
}
