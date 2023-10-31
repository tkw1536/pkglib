package lreflect

import (
	"fmt"
	"reflect"
)

func ExampleUnsafeSetAnyValue() {
	private := HasAPrivateField{}

	// get and set the private field
	value := reflect.ValueOf(&private).Elem().FieldByName("private")
	UnsafeSetAnyValue(value, reflect.ValueOf("I was set via reflect"))

	fmt.Println(private.Private())
	// Output: I was set via reflect
}
