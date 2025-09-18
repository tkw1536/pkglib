// Package reflectx provides extensions to the reflect package
//
//spellchecker:words reflectx
package reflectx_test

//spellchecker:words reflect pkglib reflectx
import (
	"fmt"
	"reflect"

	"go.tkw01536.de/pkglib/reflectx"
)

//spellchecker:words pkglib nolint

// Iterate over the fields of a struct.
func ExampleIterFields() {
	type Embed struct {
		EmbeddedField string // field in an embedded struct
	}

	type SomeStruct struct {
		Field string // regular field
		//lint:ignore U1000 // false positive: used by TypeFor below
		string         // embedded non-struct, not called recursively
		Embed          // an embed
		Another string //
	}

	for f, index := range reflectx.IterFields(reflect.TypeFor[SomeStruct]()) {
		fmt.Println("encountered field", f.Name, "with index", index)
	}

	// Output: encountered field Field with index 0
	// encountered field string with index 1
	// encountered field Embed with index 2
	// encountered field Another with index 3
}

// Iterate over the fields of a struct.
func ExampleIterAllFields() {
	type Embed struct {
		EmbeddedField string // field in an embedded struct
	}

	type SomeStruct struct {
		Field string // regular field
		//lint:ignore U1000 // false positive: used by TypeFor call below
		string         // embedded non-struct, not called recursively
		Embed          // an embed
		Another string // another field
	}

	for f, index := range reflectx.IterAllFields(reflect.TypeFor[SomeStruct]()) {
		fmt.Println("encountered field", f.Name, "with index", index)
	}

	// Output: encountered field Field with index [0]
	// encountered field string with index [1]
	// encountered field EmbeddedField with index [2 0]
	// encountered field Another with index [3]
}
