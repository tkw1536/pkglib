package perf_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tkw1536/pkglib/perf"
)

func TestSnapshot_BytesString(t *testing.T) {
	type fields struct {
		Time    time.Time
		Bytes   int64
		Objects int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"0", fields{Bytes: 0}, "0 B"},

		{"1 kB", fields{Bytes: 1024}, "1.0 kB"},
		{"1 MB", fields{Bytes: 1024 * 1024}, "1.0 MB"},
		{"124 MB", fields{Bytes: 123456789}, "124 MB"},
		{"1.1 GB", fields{Bytes: 1024 * 1024 * 1024}, "1.1 GB"},
		{"1 TB", fields{Bytes: 1024 * 1024 * 1024 * 1024}, "1.1 TB"},

		{"-1 kB", fields{Bytes: -1024}, "-1.0 kB"},
		{"-1 MB", fields{Bytes: -1024 * 1024}, "-1.0 MB"},
		{"-124 MB", fields{Bytes: -123456789}, "-124 MB"},
		{"-1.1 GB", fields{Bytes: -1024 * 1024 * 1024}, "-1.1 GB"},
		{"-1 TB", fields{Bytes: -1024 * 1024 * 1024 * 1024}, "-1.1 TB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := perf.Snapshot{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := snapshot.BytesString(); got != tt.want {
				t.Errorf("Snapshot.BytesString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnapshot_ObjectsString(t *testing.T) {
	type fields struct {
		Time    time.Time
		Bytes   int64
		Objects int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"-69", fields{Objects: -69}, "-69 objects"},
		{"-1", fields{Objects: -1}, "-1 objects"},
		{"0", fields{Objects: 0}, "0 objects"},
		{"1", fields{Objects: 1}, "1 object"},
		{"42", fields{Objects: 42}, "42 objects"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := perf.Snapshot{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := snapshot.ObjectsString(); got != tt.want {
				t.Errorf("Snapshot.ObjectsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnapshot_String(t *testing.T) {
	type fields struct {
		Time    time.Time
		Bytes   int64
		Objects int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"default", fields{Time: time.Unix(0, 0), Bytes: 1024, Objects: 13}, "1.0 kB (13 objects) used at Jan  1 00:00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot := perf.Snapshot{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := snapshot.String(); got != tt.want {
				t.Errorf("Snapshot.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleDiff() {
	// Diff holds both the amount of time an operation took,
	// the number of bytes consumed, and the total number of allocated objects.
	diff := perf.Diff{
		Time:    15 * time.Second,
		Bytes:   100,
		Objects: 100,
	}
	fmt.Println(diff)
	// Output: 15s, 100 B, 100 objects
}

func TestSnapshot_Sub(t *testing.T) {
	type fields struct {
		Time    time.Time
		Bytes   int64
		Objects int64
	}
	type args struct {
		other perf.Snapshot
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   perf.Diff
	}{
		{"identical sub", fields{Time: time.Unix(0, 0), Bytes: 1024, Objects: 13}, args{perf.Snapshot{Time: time.Unix(0, 0), Bytes: 1024, Objects: 13}}, perf.Diff{Time: 0, Bytes: 0, Objects: 0}},
		{"positive sub", fields{Time: time.Unix(1000, 0), Bytes: 2048, Objects: 69}, args{perf.Snapshot{Time: time.Unix(0, 0), Bytes: 1024, Objects: 42}}, perf.Diff{Time: 1000 * time.Second, Bytes: 1024, Objects: 27}},
		{"negative sub", fields{Time: time.Unix(0, 0), Bytes: 1024, Objects: 42}, args{perf.Snapshot{Time: time.Unix(1000, 0), Bytes: 2048, Objects: 69}}, perf.Diff{Time: -1000 * time.Second, Bytes: -1024, Objects: -27}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := perf.Snapshot{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := s.Sub(tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Snapshot.Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiff_BytesString(t *testing.T) {
	type fields struct {
		Time    time.Duration
		Bytes   int64
		Objects int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"0", fields{Bytes: 0}, "0 B"},

		{"1 kB", fields{Bytes: 1024}, "1.0 kB"},
		{"1 MB", fields{Bytes: 1024 * 1024}, "1.0 MB"},
		{"124 MB", fields{Bytes: 123456789}, "124 MB"},
		{"1.1 GB", fields{Bytes: 1024 * 1024 * 1024}, "1.1 GB"},
		{"1 TB", fields{Bytes: 1024 * 1024 * 1024 * 1024}, "1.1 TB"},

		{"-1 kB", fields{Bytes: -1024}, "-1.0 kB"},
		{"-1 MB", fields{Bytes: -1024 * 1024}, "-1.0 MB"},
		{"-124 MB", fields{Bytes: -123456789}, "-124 MB"},
		{"-1.1 GB", fields{Bytes: -1024 * 1024 * 1024}, "-1.1 GB"},
		{"-1 TB", fields{Bytes: -1024 * 1024 * 1024 * 1024}, "-1.1 TB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := perf.Diff{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := diff.BytesString(); got != tt.want {
				t.Errorf("Diff.BytesString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiff_ObjectsString(t *testing.T) {
	type fields struct {
		Time    time.Duration
		Bytes   int64
		Objects int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"-69", fields{Objects: -69}, "-69 objects"},
		{"-1", fields{Objects: -1}, "-1 objects"},
		{"0", fields{Objects: 0}, "0 objects"},
		{"1", fields{Objects: 1}, "1 object"},
		{"42", fields{Objects: 42}, "42 objects"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := perf.Diff{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := diff.ObjectsString(); got != tt.want {
				t.Errorf("Diff.ObjectsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiff_String(t *testing.T) {
	type fields struct {
		Time    time.Duration
		Bytes   int64
		Objects int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"default", fields{Time: 0, Bytes: 1024, Objects: 13}, "0s, 1.0 kB, 13 objects"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := perf.Diff{
				Time:    tt.fields.Time,
				Bytes:   tt.fields.Bytes,
				Objects: tt.fields.Objects,
			}
			if got := diff.String(); got != tt.want {
				t.Errorf("Diff.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
