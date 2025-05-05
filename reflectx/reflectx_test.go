// Package reflectx provides extensions to the reflect package
//
//spellchecker:words reflectx
package reflectx_test

//spellchecker:words reflect testing github pkglib reflectx
import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/reflectx"
)

//spellchecker:words pkglib nolint recvcheck

// counter is used for testing.
//
//nolint:recvcheck // testing code
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

	// make a cp and increment the cp
	cp, _ := reflectx.CopyInterface(original)
	cp.Inc()

	// print the value of the original counter and the copy
	// the copy is also a pointer
	fmt.Println("original counter", original)
	fmt.Println("copy of counter", cp)

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

	// make a cp and increment the cp
	cp, _ := reflectx.CopyInterface(original)
	cp.(interface{ Inc() }).Inc()

	// print the original value and the new value
	// the original is a plain value, the copy is a pointer!
	fmt.Println("original counter", original)
	fmt.Println("copy of counter", cp)

	// Output: original counter {0}
	// copy of counter &{1}
}

type TypeForTesting struct{}

func TestNameOf(t *testing.T) {
	t.Parallel()

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
			"package in standard library",
			reflect.TypeFor[reflect.Type](),
			"reflect.Type",
		},
		{
			"type in this library",
			reflect.TypeFor[TypeForTesting](),
			"github.com/tkw1536/pkglib/reflectx_test.TypeForTesting",
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
			t.Parallel()

			if got := reflectx.NameOf(tt.T); got != tt.want {
				t.Errorf("NameOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
