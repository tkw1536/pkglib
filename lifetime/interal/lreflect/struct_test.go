package lreflect

import (
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

func Test_ImplementsAsStructPointer(t *testing.T) {
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
			if got, _ := ImplementsAsStructPointer(tt.args.I, tt.args.T); got != tt.want {
				t.Errorf("ImplementsAsStructPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}
