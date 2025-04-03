// Package reflectx provides extensions to the reflect package
//
//spellchecker:words reflectx
package reflectx_test

//spellchecker:words reflect github pkglib reflectx
import (
	"fmt"
	"reflect"

	"github.com/tkw1536/pkglib/reflectx"
)

//spellchecker:words pkglib nolint

// Iterate over the fields of a struct.
func ExampleIterFields() {
	type Embed struct {
		EmbeddedField string // field in an embedded struct
	}

	//nolint:unused
	type SomeStruct struct {
		Field   string // regular field
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

	//nolint:unused
	type SomeStruct struct {
		Field   string // regular field
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
