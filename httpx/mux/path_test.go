package mux_test

import (
	"testing"

	"github.com/tkw1536/pkglib/httpx/mux"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct{ input, want string }{
		// Already clean
		{"", "/"},
		{"abc", "/abc/"},
		{"abc/def", "/abc/def/"},
		{"a/b/c", "/a/b/c/"},
		{".", "/"},
		{"..", "/"},
		{"../..", "/"},
		{"../../abc", "/abc/"},
		{"/abc", "/abc/"},
		{"/", "/"},

		// Remove trailing slash
		{"abc/", "/abc/"},
		{"abc/def/", "/abc/def/"},
		{"a/b/c/", "/a/b/c/"},
		{"./", "/"},
		{"../", "/"},
		{"../../", "/"},
		{"/abc/", "/abc/"},

		// Remove doubled slash
		{"abc//def//ghi", "/abc/def/ghi/"},
		{"//abc", "/abc/"},
		{"///abc", "/abc/"},
		{"//abc//", "/abc/"},
		{"abc//", "/abc/"},

		// Remove . elements
		{"abc/./def", "/abc/def/"},
		{"/./abc/def", "/abc/def/"},
		{"abc/.", "/abc/"},

		// Remove .. elements
		{"abc/def/ghi/../jkl", "/abc/def/jkl/"},
		{"abc/def/../ghi/../jkl", "/abc/jkl/"},
		{"abc/def/..", "/abc/"},
		{"abc/def/../..", "/"},
		{"/abc/def/../..", "/"},
		{"abc/def/../../..", "/"},
		{"/abc/def/../../..", "/"},
		{"abc/def/../../../ghi/jkl/../../../mno", "/mno/"},

		// Combinations
		{"abc/./../def", "/def/"},
		{"abc//./../def", "/def/"},
		{"abc/../../././../def", "/def/"},
	}
	// TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := mux.NormalizePath(tt.input); got != tt.want {
				t.Errorf("NormalizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
