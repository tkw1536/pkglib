package collection

import (
	"fmt"
	"strings"
)

func ExampleFirst() {
	values := []int{-1, 0, 1}
	fmt.Println(
		First(values, func(v int) bool {
			return v > 0
		}),
	)

	fmt.Println(
		First(values, func(v int) bool {
			return v > 2 // no such value exists
		}),
	)
	// Output: 1
	// 0
}

func ExampleAny() {
	values := []int{-1, 0, 1}
	fmt.Println(
		Any(values, func(v int) bool {
			return v > 0
		}),
	)

	fmt.Println(
		Any(values, func(v int) bool {
			return v > 2 // no such value exists
		}),
	)
	// Output: true
	// false
}

func ExampleMapSlice() {
	values := []int{-1, 0, 1}
	fmt.Println(
		MapSlice(values, func(v int) float64 {
			return float64(v) / 2
		}),
	)

	// Output: [-0.5 0 0.5]
}

func ExamplePartition() {
	values := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(
		Partition(values, func(v int) int {
			return v % 3
		}),
	)

	// Output: map[0:[0 3 6 9] 1:[1 4 7 10] 2:[2 5 8]]
}

func ExampleNonSequential() {
	values := []string{"a", "aa", "b", "bb", "c"}
	fmt.Println(
		NonSequential(values, func(prev, current string) bool {
			return strings.HasPrefix(current, prev)
		}),
	)
	// Output: [a b c]
}
