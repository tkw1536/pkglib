//spellchecker:words collection
package collection

//spellchecker:words reflect strings testing slices
import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"slices"
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
	t.Parallel()

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
			t.Parallel()

			if got := Deduplicate(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeepFunc(t *testing.T) {
	t.Parallel()

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
		{"KeepFunc on value", args{[]string{"a", "b", "c"}, func(s string) bool { return s == "a" }}, []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := KeepFunc(slices.Clone(tt.args.s), tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeepFunc() = %v, want %v", got, tt.want)
			}
			if got := KeepFuncClone(slices.Clone(tt.args.s), tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeepFuncClone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleKeepFunc() {
	values := []string{"hello", "world", "how", "are", "you"}
	valuesF := KeepFunc(values, func(s string) bool {
		return s != "how"
	})
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "are", "you", ""}
	// []string{"hello", "world", "are", "you"}
}

func ExampleKeepFunc_allTrue() {
	values := []string{"hello", "world", "how", "are", "you"}
	valuesF := KeepFunc(values, func(s string) bool { return true })
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "how", "are", "you"}
	// []string{"hello", "world", "how", "are", "you"}
}

func ExampleKeepFunc_allFalse() {
	values := []string{"hello", "world", "how", "are", "you"}
	valuesF := KeepFunc(values, func(s string) bool { return false })
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"", "", "", "", ""}
	// []string{}
}

func ExampleKeepFuncClone() {
	values := []string{"hello", "world", "how", "are", "you"}
	valuesF := KeepFuncClone(values, func(s string) bool {
		return s != "how"
	})
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "how", "are", "you"}
	// []string{"hello", "world", "are", "you"}
}

func ExampleKeepFuncClone_order() {
	values := []string{"hello", "world", "how", "are", "you"}

	// the KeepFunc function is guaranteed to be called in order
	index := 0
	valuesF := KeepFuncClone(values, func(s string) bool {
		// KeepFunc every even element
		res := index%2 == 0
		index++
		return res
	})
	fmt.Printf("%#v\n%#v\n", values, valuesF)
	// Output: []string{"hello", "world", "how", "are", "you"}
	// []string{"hello", "how", "you"}
}
