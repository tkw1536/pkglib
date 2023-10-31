package lreflect

import (
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/reflectx"
)

func Test_ImplementsAsSliceInterface(t *testing.T) {
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
			if got, _ := ImplementsAsSliceInterface(tt.args.I, tt.args.T); got != tt.want {
				t.Errorf("ImplementsAsSliceInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FilterSliceInterface(t *testing.T) {

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
			if got, _ := FilterSliceInterface(reflect.ValueOf(tt.args.S), tt.args.I); !reflect.DeepEqual(got.Interface(), tt.want) {
				t.Errorf("FilterSliceInterface() = %#v, want %#v", got, tt.want)
			}
		})
	}
}