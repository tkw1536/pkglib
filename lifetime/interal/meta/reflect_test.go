// Package meta provides metadata methods for the dependency injection framework
package meta

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/reflectx"
)

// HasAPrivateField has a private field
type HasAPrivateField struct {
	private string
}

func (hp HasAPrivateField) Private() string {
	return hp.private
}

func ExampleunsafeSetAnyValue() {
	private := HasAPrivateField{}

	// get and set the private field
	value := reflect.ValueOf(&private).Elem().FieldByName("private")
	unsafeSetAnyValue(value, reflect.ValueOf("I was set via reflect"))

	fmt.Println(private.Private())
	// Output: I was set via reflect
}

// SomeInterface provides two Methods
type SomeInterface interface {
	MethodA()
	MethodB()
}

// OtherInterface provides a single Method
// And SomeInterface is a superset of OtherInterface.
type OtherInterface interface{ MethodA() }

// SomeStruct implements SomeInterface
type SomeStruct struct {
	Value int
}

func (SomeStruct) MethodA() {}
func (SomeStruct) MethodB() {}

// OtherStruct implements OtherInterface
type OtherStruct struct {
	Value int
}

func (OtherStruct) MethodA() {}

func Test_implementsAsStructPointer(t *testing.T) {
	type args struct {
		I reflect.Type
		T reflect.Type
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"implements method as a pointer",
			args{
				reflectx.TypeFor[OtherInterface](),
				reflectx.TypeFor[*OtherStruct](),
			},
			true,
		},
		{
			"does not implement method",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[*OtherStruct](),
			},
			false,
		},
		{
			"implements but not as pointer",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[SomeStruct](),
			},
			false,
		},
		{
			"implements but not a struct",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[SomeInterface](),
			},
			false,
		},
		{
			"non-interface passed",
			args{
				reflectx.TypeFor[SomeStruct](),
				reflectx.TypeFor[SomeInterface](),
			},
			false,
		},
		{
			"nil interface passed",
			args{
				nil,
				reflectx.TypeFor[*SomeStruct](),
			},
			false,
		},
		{
			"nil struct passed",
			args{
				reflectx.TypeFor[SomeInterface](),
				nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := implementsAsStructPointer(tt.args.I, tt.args.T); got != tt.want {
				t.Errorf("implementsAsStructPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_implementsAsSliceInterface(t *testing.T) {
	type args struct {
		I reflect.Type
		T reflect.Type
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"interface slice implements",
			args{
				reflectx.TypeFor[OtherInterface](),
				reflectx.TypeFor[[]SomeInterface](),
			},
			true,
		},
		{
			"struct slice does not implement",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[[]OtherStruct](),
			},
			false,
		},
		{
			"struct slice does implement but isn't an interface",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[[]SomeStruct](),
			},
			false,
		},
		{
			"non-slice implementing struct passed",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[*SomeStruct](),
			},
			false,
		},
		{
			"non-slice non-implementing struct passed",
			args{
				reflectx.TypeFor[SomeInterface](),
				reflectx.TypeFor[*OtherStruct](),
			},
			false,
		},
		{
			"nil interface passed",
			args{
				nil,
				reflectx.TypeFor[[]*SomeStruct](),
			},
			false,
		},
		{
			"nil struct passed",
			args{
				reflectx.TypeFor[SomeInterface](),
				nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := implementsAsSliceInterface(tt.args.I, tt.args.T); got != tt.want {
				t.Errorf("implementsAsSliceInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterSliceInterface(t *testing.T) {

	somestruct := func(i int) *SomeStruct {
		v := SomeStruct{Value: i}
		return &v
	}
	otherstruct := func(i int) *OtherStruct {
		v := OtherStruct{Value: i}
		return &v
	}

	type args struct {
		S any
		I reflect.Type
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			"filter slice type",
			args{
				[]any{
					somestruct(1),
					otherstruct(2),
					otherstruct(3),
					somestruct(4),
				},
				reflectx.TypeFor[SomeInterface](),
			},
			[]SomeInterface{somestruct(1), somestruct(4)},
		},
		{
			"slice with no matching elements",
			args{
				[]any{
					otherstruct(1),
					otherstruct(2),
				},
				reflectx.TypeFor[SomeInterface](),
			},
			[]SomeInterface{},
		},
		{
			"non-interface passed",
			args{
				[]any{
					otherstruct(1),
					otherstruct(2),
				},
				reflectx.TypeFor[string](),
			},
			[]string(nil),
		},
		{
			"non-slice passed",
			args{
				"hello world",
				reflectx.TypeFor[SomeInterface](),
			},
			[]SomeInterface(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterSliceInterface(reflect.ValueOf(tt.args.S), tt.args.I); !reflect.DeepEqual(got.Interface(), tt.want) {
				t.Errorf("filterSliceInterface() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
