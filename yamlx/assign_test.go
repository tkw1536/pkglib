//spellchecker:words yamlx
package yamlx_test

//spellchecker:words testing pkglib yamlx gopkg yaml
import (
	"testing"

	"go.tkw01536.de/pkglib/yamlx"
	"gopkg.in/yaml.v3"
)

func TestAssign(t *testing.T) {
	t.Parallel()

	type args struct {
		yamlContent string
		key         string
		value       any
	}
	tests := []struct {
		name        string
		args        args
		wantErr     string
		wantYAML    string
		mappingNode func(t *testing.T, node *yaml.Node) *yaml.Node
	}{
		{
			name: "update existing key",
			args: args{
				yamlContent: "a: 0\nb: 1",
				key:         "b",
				value:       2,
			},
			wantYAML:    "a: 0\nb: 2\n",
			mappingNode: rootMapping,
		},
		{
			name: "append new key",
			args: args{
				yamlContent: "a: 0",
				key:         "b",
				value:       1,
			},
			wantYAML:    "a: 0\nb: 1\n",
			mappingNode: rootMapping,
		},
		{
			name: "assign nested mapping",
			args: args{
				yamlContent: "a: 0",
				key:         "b",
				value: map[string]any{
					"c": 2,
					"d": 3,
				},
			},
			wantYAML:    "a: 0\nb:\n    c: 2\n    d: 3\n",
			mappingNode: rootMapping,
		},
		{
			name: "replace scalar with sequence",
			args: args{
				yamlContent: "a: 0",
				key:         "a",
				value:       []int{1, 2, 3},
			},
			wantYAML:    "a:\n    - 1\n    - 2\n    - 3\n",
			mappingNode: rootMapping,
		},
		{
			name: "assign on non-mapping node",
			args: args{
				yamlContent: "a: 0",
				key:         "a",
				value:       1,
			},
			wantErr: "expected mapping node",
			mappingNode: func(t *testing.T, node *yaml.Node) *yaml.Node {
				t.Helper()
				value, err := yamlx.Find(node, "a")
				if err != nil {
					t.Fatalf("Find() error = %v", err)
				}
				return value
			},
		},
		{
			name: "assign on nil node",
			args: args{
				key:   "a",
				value: 1,
			},
			wantErr: "expected mapping node",
			mappingNode: func(t *testing.T, node *yaml.Node) *yaml.Node {
				t.Helper()
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var document *yaml.Node
			if tt.args.yamlContent != "" {
				document = mustUnmarshal(t, tt.args.yamlContent)
			}

			mapping := tt.mappingNode(t, document)

			valueNode, err := yamlx.Marshal(tt.args.value)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			err = yamlx.Assign(mapping, tt.args.key, *valueNode)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tt.wantErr {
				t.Errorf("Assign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == "" {
				got := mustMarshal(t, mapping)
				if got != tt.wantYAML {
					t.Errorf("Assign() result =\n%s\nwant =\n%s", got, tt.wantYAML)
				}
			}
		})
	}
}

func rootMapping(t *testing.T, node *yaml.Node) *yaml.Node {
	t.Helper()
	if node == nil || len(node.Content) == 0 {
		t.Fatal("expected document with content")
	}
	return node.Content[0]
}
