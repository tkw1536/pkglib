//spellchecker:words lreflect
package lreflect_test

//spellchecker:words reflect github pkglib lifetime interal lreflect
import (
	"fmt"
	"reflect"

	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
)

func ExampleUnsafeSetAnyValue() {
	private := HasAPrivateField{}

	// get and set the private field
	value := reflect.ValueOf(&private).Elem().FieldByName("private")
	_ = lreflect.UnsafeSetAnyValue(value, reflect.ValueOf("I was set via reflect"))

	fmt.Println(private.Private())
	// Output: I was set via reflect
}
