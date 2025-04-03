//spellchecker:words lreflect
package lreflect_test

//spellchecker:words reflect testing github pkglib lifetime interal lreflect
import (
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/lifetime/interal/lreflect"
)

func Test_ImplementsAsSliceInterface(t *testing.T) {
	t.Parallel()

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
				reflect.TypeFor[OtherInterface](),
				reflect.TypeFor[[]SomeInterface](),
			},
			true,
		},
		{
			"struct slice does not implement",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[[]OtherStruct](),
			},
			false,
		},
		{
			"struct slice does implement but isn't an interface",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[[]SomeStruct](),
			},
			false,
		},
		{
			"non-slice implementing struct passed",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[*SomeStruct](),
			},
			false,
		},
		{
			"non-slice non-implementing struct passed",
			args{
				reflect.TypeFor[SomeInterface](),
				reflect.TypeFor[*OtherStruct](),
			},
			false,
		},
		{
			"nil interface passed",
			args{
				nil,
				reflect.TypeFor[[]*SomeStruct](),
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
			t.Parallel()

			if got, _ := lreflect.ImplementsAsSliceInterface(tt.args.I, tt.args.T); got != tt.want {
				t.Errorf("ImplementsAsSliceInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FilterSliceInterface(t *testing.T) {
	t.Parallel()

	someStruct := func(i int) *SomeStruct {
		v := SomeStruct{Value: i}
		return &v
	}
	otherStruct := func(i int) *OtherStruct {
		v := OtherStruct{Value: i}
		return &v
	}

	type args struct {
		S reflect.Value
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
				reflect.ValueOf(
					[]any{
						someStruct(1),
						otherStruct(2),
						otherStruct(3),
						someStruct(4),
					},
				),
				reflect.TypeFor[SomeInterface](),
			},
			[]SomeInterface{someStruct(1), someStruct(4)},
		},
		{
			"slice with no matching elements",
			args{
				reflect.ValueOf(
					[]any{
						otherStruct(1),
						otherStruct(2),
					},
				),
				reflect.TypeFor[SomeInterface](),
			},
			[]SomeInterface{},
		},
		{
			"non-interface passed",
			args{
				reflect.ValueOf(
					[]any{
						otherStruct(1),
						otherStruct(2),
					},
				),
				reflect.TypeFor[string](),
			},
			nil,
		},
		{
			"non-slice passed",
			args{
				reflect.ValueOf("hello world"),
				reflect.TypeFor[SomeInterface](),
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := lreflect.FilterSliceInterface(tt.args.S, tt.args.I)
			var gotActual any
			if err == nil {
				gotActual = got.Interface()
			}
			if !reflect.DeepEqual(gotActual, tt.want) {
				t.Errorf("FilterSliceInterface() = %#v, want %#v", gotActual, tt.want)
			}
		})
	}
}
