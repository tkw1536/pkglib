//spellchecker:words lreflect
package lreflect

//spellchecker:words reflect testing
import (
	"reflect"
	"testing"
)

// HasAPrivateField has a private field.
type HasAPrivateField struct {
	private string
}

func (hp HasAPrivateField) Private() string {
	return hp.private
}

// SomeInterface provides two Methods.
type SomeInterface interface {
	MethodA()
	MethodB()
}

// OtherInterface provides a single Method
// And SomeInterface is a superset of OtherInterface.
type OtherInterface interface{ MethodA() }

// SomeStruct implements SomeInterface.
type SomeStruct struct {
	Value int
}

func (SomeStruct) MethodA() {}
func (SomeStruct) MethodB() {}

// OtherStruct implements OtherInterface.
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
				reflect.TypeFor[OtherInterface](),
				reflect.TypeFor[*OtherStruct](),
			},
			true,
		},
		{
			"does not implement method",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[*OtherStruct](),
			},
			false,
		},
		{
			"implements but not as pointer",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[SomeStruct](),
			},
			false,
		},
		{
			"implements but not a struct",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[SomeInterface](),
			},
			false,
		},
		{
			"non-interface passed",
			args{
				reflect.TypeFor[SomeStruct](),
				reflect.TypeFor[SomeInterface](),
			},
			false,
		},
		{
			"nil interface passed",
			args{
				nil,
				reflect.TypeFor[*SomeStruct](),
			},
			false,
		},
		{
			"nil struct passed",
			args{
				reflect.TypeFor[SomeInterface](),
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
