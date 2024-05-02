// Package reflectx provides extensions to the reflect package
//
//spellchecker:words reflectx
package reflectx

//spellchecker:words reflect testing
import (
	"fmt"
	"reflect"
	"testing"
)

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

	// prevent unused field warning
	var s SomeStruct
	_ = s.string

	fmt.Println(
		"returned:",
		IterateFields(reflect.TypeFor[SomeStruct](), func(f reflect.StructField, index int) (stop bool) {
			fmt.Println("encountered field", f.Name, "with index", index)
			return false // do not stop
		}),
	)

	// Output: encountered field Field with index 0
	// encountered field string with index 1
	// encountered field Embed with index 2
	// encountered field Another with index 3
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

	// prevent unused field warning
	var s SomeStruct
	_ = s.string

	fmt.Println(
		"returned:",
		IterateAllFields(reflect.TypeFor[SomeStruct](), func(f reflect.StructField, index ...int) (stop bool) {
			fmt.Println("encountered field", f.Name, "with index", index)
			return false // do not stop
		}),
	)

	// Output: encountered field Field with index [0]
	// encountered field string with index [1]
	// encountered field EmbeddedField with index [2 0]
	// encountered field Another with index [3]
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

	// prevent unused field warning
	var s SomeStruct
	_ = s.string

	fmt.Println(
		"returned:",
		IterateAllFields(reflect.TypeFor[SomeStruct](), func(f reflect.StructField, index ...int) (cancel bool) {
			fmt.Println("encountered field", f.Name, "with index", index)
			return f.Name == "EmbeddedField" // cancel on embedded field
		}),
	)

	// Output: encountered field Field with index [0]
	// encountered field string with index [1]
	// encountered field EmbeddedField with index [2 0]
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

func ExampleCopyInterface_pointer() {
	// Inc is an interface that increments
	type Inc interface {
		Inc()
	}

	// make a pointer to a counter
	original := Inc(&counter{Value: 0})

	// make a copy and increment the copy
	copy, _ := CopyInterface(original)
	copy.Inc()

	// print the value of the original counter and the copy
	// the copy is also a pointer
	fmt.Println("original counter", original)
	fmt.Println("copy of counter", copy)

	// Output: original counter &{0}
	// copy of counter &{1}
}

func ExampleCopyInterface_lift() {

	// AsInt is an interface that returns an integer value
	type AsInt interface {
		AsInt() int
	}

	// make a *non-pointer* counter
	original := AsInt(counter{Value: 0})

	// make a copy and increment the copy
	copy, _ := CopyInterface(original)
	copy.(interface{ Inc() }).Inc()

	// print the original value and the new value
	// the original is a plain value, the copy is a pointer!
	fmt.Println("original counter", original)
	fmt.Println("copy of counter", copy)

	// Output: original counter {0}
	// copy of counter &{1}
}

type TypeForTesting struct{}

func TestNameOf(t *testing.T) {
	funcForTesting := func() {}

	tests := []struct {
		name string
		T    reflect.Type
		want string
	}{
		{
			"built-in type",
			reflect.TypeFor[string](),
			".string",
		},
		{
			"package in standard libary",
			reflect.TypeFor[reflect.Type](),
			"reflect.Type",
		},
		{
			"type in this library",
			reflect.TypeFor[TypeForTesting](),
			"github.com/tkw1536/pkglib/reflectx.TypeForTesting",
		},

		{
			"non-named type (pointer)",
			reflect.TypeFor[*string](),
			"",
		},
		{
			"non-named type (slice)",
			reflect.TypeFor[[]string](),
			"",
		},
		{
			"non-named type (map)",
			reflect.TypeFor[map[string]string](),
			"",
		},
		{
			"non-named type (local function)",
			reflect.TypeOf(funcForTesting),
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NameOf(tt.T); got != tt.want {
				t.Errorf("NameOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
