package collection

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
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

func TestDeduplicate(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"nil slice", args{nil}, nil},
		{"no duplicates", args{[]string{"a", "b", "c", "d"}}, []string{"a", "b", "c", "d"}},
		{"some duplicates", args{[]string{"b", "c", "c", "d", "a", "b", "c", "d"}}, []string{"b", "c", "d", "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Deduplicate(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicates() = %v, want %v", got, tt.want)
			}
		})
	}
}
