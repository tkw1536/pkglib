package collection

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/exp/slices"
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

func TestFilter(t *testing.T) {
	type args struct {
		s []string
		f func(string) bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"nil slice", args{nil, func(string) bool { panic("never called") }}, nil},
		{"empty slice", args{[]string{}, func(string) bool { panic("never called") }}, []string{}},
		{"filter on value", args{[]string{"a", "b", "c"}, func(s string) bool { return s == "a" }}, []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(slices.Clone(tt.args.s), tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
			if got := FilterClone(slices.Clone(tt.args.s), tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterClone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleFilter() {
	values := []string{"hello", "world", "how", "are", "you"}
	valuesF := Filter(values, func(s string) bool {
		return s != "how"
	})
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "are", "you", ""}
	// []string{"hello", "world", "are", "you"}
}

func ExampleFilterClone() {
	values := []string{"hello", "world", "how", "are", "you"}
	valuesF := FilterClone(values, func(s string) bool {
		return s != "how"
	})
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "how", "are", "you"}
	// []string{"hello", "world", "are", "you"}
}

func ExampleFilterClone_order() {
	values := []string{"hello", "world", "how", "are", "you"}

	// the Filter function is guaranteed to be called in order
	index := 0
	valuesF := FilterClone(values, func(s string) bool {
		// filter every even element
		res := index%2 == 0
		index++
		return res
	})
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "how", "are", "you"}
	// []string{"hello", "how", "you"}
}
