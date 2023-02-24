// Package reflectx provides extensions to the reflect package
package reflectx

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTypeOf(t *testing.T) {
	tests := []struct {
		name string
		got  reflect.Type
		want reflect.Type
	}{
		{
			"string",
			TypeOf[string](),
			reflect.TypeOf(string("")),
		},
		{
			"int",
			TypeOf[int](),
			reflect.TypeOf(int(0)),
		},
		{
			"slice",
			TypeOf[[]string](),
			reflect.TypeOf([]string(nil)),
		},
		{
			"array",
			TypeOf[[0]string](),
			reflect.TypeOf([0]string{}),
		},
		{
			"chan",
			TypeOf[chan string](),
			reflect.TypeOf((chan string)(nil)),
		},
		{
			"func",
			TypeOf[func(string) string](),
			reflect.TypeOf(func(string) string { return "" }),
		},
		{
			"map",
			TypeOf[map[string]string](),
			reflect.TypeOf(map[string]string(nil)),
		},
		{
			"struct",
			TypeOf[struct{ Thing string }](),
			reflect.TypeOf(struct{ Thing string }{}),
		},
		{
			"pointer",
			TypeOf[*struct{ Thing string }](),
			reflect.TypeOf(&struct{ Thing string }{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("TypeOf() = %v, want %v", tt.got, tt.want)
			}
		})
	}
}

// Iterate over the fields of a struct
func ExampleIterateFields() {

	type Embed struct {
		EmbeddedField string // field in an embedded struct
	}

	type SomeStruct struct {
		Field   string // regular field
		string         // embedded non-struct, not called recursively
		Embed          // an embed
		Another string //
	}

	fmt.Println(
		"returned:",
		IterateFields(TypeOf[SomeStruct](), func(f reflect.StructField, index int) (stop bool) {
			fmt.Println("encounted field", f.Name, "with index", index)
			return false // do not stop
		}),
	)

	// Output: encounted field Field with index 0
	// encounted field string with index 1
	// encounted field Embed with index 2
	// encounted field Another with index 3
	// returned: false
}

// Iterate over the fields of a struct
func ExampleIterateAllFields() {

	type Embed struct {
		EmbeddedField string // field in an embedded struct
	}

	type SomeStruct struct {
		Field   string // regular field
		string         // embedded non-struct, not called recursively
		Embed          // an embed
		Another string //
	}

	fmt.Println(
		"returned:",
		IterateAllFields(TypeOf[SomeStruct](), func(f reflect.StructField, index ...int) (stop bool) {
			fmt.Println("encounted field", f.Name, "with index", index)
			return false // do not stop
		}),
	)

	// Output: encounted field Field with index [0]
	// encounted field string with index [1]
	// encounted field EmbeddedField with index [2 0]
	// encounted field Another with index [3]
	// returned: false
}

// stop the iteration over a subset of fields only
func ExampleIterateAllFields_cancel() {

	type Embed struct {
		EmbeddedField string // field in an embedded struct
	}

	type SomeStruct struct {
		Field   string // regular field
		string         // embedded non-struct, not called recursively
		Embed          // an embed
		Another string //
	}

	fmt.Println(
		"returned:",
		IterateAllFields(TypeOf[SomeStruct](), func(f reflect.StructField, index ...int) (cancel bool) {
			fmt.Println("encounted field", f.Name, "with index", index)
			return f.Name == "EmbeddedField" // cancel on embedded field
		}),
	)

	// Output: encounted field Field with index [0]
	// encounted field string with index [1]
	// encounted field EmbeddedField with index [2 0]
	// returned: true
}

// counter is used for testing
type counter struct {
	Value int
}

func (c *counter) Inc() {
	c.Value++
}

func (c counter) AsInt() int {
	return c.Value
}

func ExampleMakePointerCopy_pointer() {
	// Inc is an interface that incremenets
	type Inc interface {
		Inc()
	}

	// make a pointer to a counter
	original := Inc(&counter{Value: 0})

	// make a copy and increment the copy
	copy, _ := MakePointerCopy(original)
	copy.Inc()

	// print the value of the original counter and the copy
	// the copy is also a pointer
	fmt.Println("original counter", original)
	fmt.Println("copy of counter", copy)

	// Output: original counter &{0}
	// copy of counter &{1}
}

func ExampleMakePointerCopy_lift() {

	// AsInt is an interface that returns an integer value
	type AsInt interface {
		AsInt() int
	}

	// make a *non-pointer* counter
	original := AsInt(counter{Value: 0})

	// make a copy and increment the copy
	copy, _ := MakePointerCopy(original)
	copy.(interface{ Inc() }).Inc()

	// print the original value and the new value
	// the original is a plain value, the copy is a pointer!
	fmt.Println("original counter", original)
	fmt.Println("copy of counter", copy)

	// Output: original counter {0}
	// copy of counter &{1}
}

func ExampleCopy() {
	// counter is a struct holding a single number

	// copying a data structure directly
	func() {
		original := counter{Value: 0}
		copy := Copy(original, false)
		copy.Value++
		fmt.Println("original data", original)
		fmt.Println("copy of data", copy)
	}()

	// copy a pointed value, which copies only the pointer
	func() {
		original := &counter{Value: 0}
		copy := Copy(original, false)
		copy.Value++
		fmt.Println("original pointer", original)
		fmt.Println("copy of pointer", copy)
	}()

	// copy a pointed value with element which copies the data
	func() {
		original := &counter{Value: 0}
		copy := Copy(original, true)
		copy.Value++
		fmt.Println("original pointer", original)
		fmt.Println("copied data", copy)
	}()

	// Output: original data {0}
	// copy of data {1}
	// original pointer &{1}
	// copy of pointer &{1}
	// original pointer &{0}
	// copied data &{1}
}
