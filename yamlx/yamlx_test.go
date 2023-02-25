package yamlx

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func mustMarshal(node *yaml.Node) string {
	result, err := yaml.Marshal(node)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func ExampleMarshal() {
	// take a random value to encode
	value := map[string]any{
		"count": 2,
		"numbers": map[string]any{
			"42": "the answer",
			"69": "nice",
		},
	}

	// marshal it as a node
	node, err := Marshal(value)
	if err != nil {
		panic(err)
	}

	// and print it out
	fmt.Println(mustMarshal(node))

	// Output: count: 2
	// numbers:
	//     "42": the answer
	//     "69": nice
}

var exampleYAML *yaml.Node

func init() {
	exampleYAML = new(yaml.Node)
	err := yaml.Unmarshal([]byte(`
a: 0
b:
    c: 2
    d: 3
    e:
        f: 5
g: 6
h:
    - 7
    - 8
`), exampleYAML)
	if err != nil {
		panic(err)
	}
}

func TestFind(t *testing.T) {
	type args struct {
		node *yaml.Node
		path []string
	}
	tests := []struct {
		name    string
		args    args
		want    *yaml.Node
		wantErr string
	}{
		{
			name:    "get root node",
			args:    args{node: exampleYAML, path: nil},
			want:    exampleYAML.Content[0],
			wantErr: "",
		},

		{
			name:    "get level-1 node",
			args:    args{node: exampleYAML, path: []string{"a"}},
			want:    exampleYAML.Content[0].Content[1],
			wantErr: "",
		},

		{
			name:    "get non-existent level-1 node",
			args:    args{node: exampleYAML, path: []string{"i"}},
			want:    nil,
			wantErr: `child not found: "i"`,
		},
		{
			name:    "get non-existent level-3 node",
			args:    args{node: exampleYAML, path: []string{"b", "e", "g"}},
			want:    nil,
			wantErr: `Node "b": Node "e": child not found: "g"`,
		},

		{
			name:    "get existent level 2 node",
			args:    args{node: exampleYAML, path: []string{"b", "c"}},
			want:    exampleYAML.Content[0].Content[3].Content[1],
			wantErr: "",
		},

		{
			name:    "get non-existent level 2 node (unexpected scalar)",
			args:    args{node: exampleYAML, path: []string{"a", "b"}},
			want:    nil,
			wantErr: `Node "a": unexpected scalar node`,
		},
		{
			name:    "get non-existent level 2 node (unexpected sequence)",
			args:    args{node: exampleYAML, path: []string{"h", "nothing"}},
			want:    nil,
			wantErr: `Node "h": unexpected sequence node`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Find(tt.args.node, tt.args.path...)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
