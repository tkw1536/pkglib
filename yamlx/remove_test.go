package yamlx_test

import (
	"testing"

	"go.tkw01536.de/pkglib/yamlx"
)

func TestRemove(t *testing.T) {
	t.Parallel()

	type args struct {
		yamlContent string
		path        []string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  string
		wantYAML string
	}{
		{
			name: "remove level-1 node",
			args: args{
				yamlContent: "a: 0\nb: 1\nc: 2",
				path:        []string{"b"},
			},
			wantErr:  "",
			wantYAML: "a: 0\nc: 2\n",
		},
		{
			name: "remove level-2 node",
			args: args{
				yamlContent: "a: 0\nb:\n    c: 2\n    d: 3\ne: 4",
				path:        []string{"b", "c"},
			},
			wantErr:  "",
			wantYAML: "a: 0\nb:\n    d: 3\ne: 4\n",
		},
		{
			name: "remove level-3 node",
			args: args{
				yamlContent: "a:\n    b:\n        c: 1\n        d: 2\n    e: 3",
				path:        []string{"a", "b", "d"},
			},
			wantErr:  "",
			wantYAML: "a:\n    b:\n        c: 1\n    e: 3\n",
		},
		{
			name: "remove non-existent level-1 node",
			args: args{
				yamlContent: "a: 0\nb: 1",
				path:        []string{"noExist"},
			},
			wantErr:  `child not found: "noExist"`,
			wantYAML: "",
		},
		{
			name: "remove non-existent level-2 node",
			args: args{
				yamlContent: "a: 0\nb:\n    c: 2",
				path:        []string{"b", "noExist"},
			},
			wantErr:  `child not found: "noExist"`,
			wantYAML: "",
		},
		{
			name: "remove from non-existent parent",
			args: args{
				yamlContent: "a: 0\nb: 1",
				path:        []string{"noExist", "child"},
			},
			wantErr:  `child not found: "noExist"`,
			wantYAML: "",
		},
		{
			name: "remove with zero-length path",
			args: args{
				yamlContent: "a: 0\nb: 1",
				path:        []string{},
			},
			wantErr:  "path is empty",
			wantYAML: "",
		},
		{
			name: "remove from scalar node",
			args: args{
				yamlContent: "a: 0\nb: 1",
				path:        []string{"a", "child"},
			},
			wantErr:  "expected mapping node",
			wantYAML: "",
		},
		{
			name: "remove from sequence node",
			args: args{
				yamlContent: "a:\n    - 1\n    - 2\nb: 3",
				path:        []string{"a", "child"},
			},
			wantErr:  "expected mapping node",
			wantYAML: "",
		},
		{
			name: "remove node via alias",
			args: args{
				yamlContent: "a: &ref\n    b: 1\n    c: 2\nd: *ref",
				path:        []string{"d", "b"},
			},
			wantErr:  "",
			wantYAML: "a: &ref\n    c: 2\nd: *ref\n",
		},
		{
			name: "remove from nested structure with aliases",
			args: args{
				yamlContent: "base: &base\n    x: 1\n    y: 2\nderived:\n    <<: *base\n    z: 3",
				path:        []string{"derived", "z"},
			},
			wantErr:  "",
			wantYAML: "base: &base\n    x: 1\n    y: 2\nderived:\n    !!merge <<: *base\n",
		},
		{
			name: "remove only remaining child",
			args: args{
				yamlContent: "a:\n    b: 1",
				path:        []string{"a", "b"},
			},
			wantErr:  "",
			wantYAML: "a: {}\n",
		},
		{
			name: "remove first of multiple children",
			args: args{
				yamlContent: "a: 1\nb: 2\nc: 3",
				path:        []string{"a"},
			},
			wantErr:  "",
			wantYAML: "b: 2\nc: 3\n",
		},
		{
			name: "remove last of multiple children",
			args: args{
				yamlContent: "a: 1\nb: 2\nc: 3",
				path:        []string{"c"},
			},
			wantErr:  "",
			wantYAML: "a: 1\nb: 2\n",
		},
		{
			name: "remove middle of multiple children",
			args: args{
				yamlContent: "a: 1\nb: 2\nc: 3",
				path:        []string{"b"},
			},
			wantErr:  "",
			wantYAML: "a: 1\nc: 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node := mustUnmarshal(t, tt.args.yamlContent)

			err := yamlx.Remove(node, tt.args.path...)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Only check the resulting YAML if no error was expected
			if tt.wantErr == "" {
				got := mustMarshal(t, node)
				if got != tt.wantYAML {
					t.Errorf("Remove() result =\n%s\nwant =\n%s", got, tt.wantYAML)
				}
			}
		})
	}
}
