// Package exit defines exit and error behavior of programs and commands.
//
//spellchecker:words exit
package exit_test

//spellchecker:words math testing pkglib exit
import (
	"math"
	"testing"

	"go.tkw01536.de/pkglib/exit"
)

func TestCode_inbounds(t *testing.T) {
	t.Parallel()

	for i := range math.MaxUint8 {
		want := exit.ExitCode(i) // #nosec G115 // in bounds by loop condition
		if got := exit.Code(i); got != want {
			t.Errorf("Code(%d) = %v, want %v", i, got, want)
		}
	}
}

func TestCode_outbounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
		want exit.ExitCode
	}{
		{"negative number", -1, math.MaxUint8},
		{"too big positive number", math.MaxUint8 + 1, math.MaxUint8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := exit.Code(tt.code); got != tt.want {
				t.Errorf("Code() = %v, want %v", got, tt.want)
			}
		})
	}
}
