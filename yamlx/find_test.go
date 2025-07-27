//spellchecker:words yamlx
package yamlx_test

//spellchecker:words reflect testing pkglib yamlx gopkg yaml
import (
	"reflect"
	"testing"

	"go.tkw01536.de/pkglib/yamlx"
	"gopkg.in/yaml.v3"
)

//spellchecker:words bref

func TestFind(t *testing.T) {
	t.Parallel()

	node := mustUnmarshal(t, "a: 0\nb: &bref\n    c: 2\n    d: 3\n    e:\n        f: 5\ng: 6\nh:\n    - 7\n    - 8\ni: *bref\n")

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
			args:    args{node: node, path: nil},
			want:    node.Content[0],
			wantErr: "",
		},

		{
			name:    "get level-1 node",
			args:    args{node: node, path: []string{"a"}},
			want:    node.Content[0].Content[1],
			wantErr: "",
		},

		{
			name:    "get non-existent level-1 node",
			args:    args{node: node, path: []string{"noExist"}},
			want:    nil,
			wantErr: `child not found: "noExist"`,
		},
		{
			name:    "get non-existent level-3 node",
			args:    args{node: node, path: []string{"b", "e", "g"}},
			want:    nil,
			wantErr: `Node "b": Node "e": child not found: "g"`,
		},

		{
			name:    "get existent level 2 node",
			args:    args{node: node, path: []string{"b", "c"}},
			want:    node.Content[0].Content[3].Content[1],
			wantErr: "",
		},

		{
			name:    "get non-existent level 2 node (unexpected scalar)",
			args:    args{node: node, path: []string{"a", "b"}},
			want:    nil,
			wantErr: `Node "a": unexpected scalar node`,
		},
		{
			name:    "get non-existent level 2 node (unexpected sequence)",
			args:    args{node: node, path: []string{"h", "nothing"}},
			want:    nil,
			wantErr: `Node "h": unexpected sequence node`,
		},

		{
			name:    "get alias level-1 node",
			args:    args{node: node, path: []string{"i"}},
			want:    node.Content[0].Content[3],
			wantErr: "",
		},
		{
			name:    "get level-2 node via alias",
			args:    args{node: node, path: []string{"i", "c"}},
			want:    node.Content[0].Content[3].Content[1],
			wantErr: "",
		},
		{
			name:    "get non-existing level-2 node via alias",
			args:    args{node: node, path: []string{"i", "noExist"}},
			want:    nil,
			wantErr: `Node "i": child not found: "noExist"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := yamlx.Find(tt.args.node, tt.args.path...)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestFind_extends(t *testing.T) {
	t.Parallel()

	node := mustUnmarshal(t, `
# some anchor nodes used for testing
red: &red
    got: "red"
green: &green
    got: "green"

# for each of the tests got should equal want
test_a:
    got: "blue"
    want: "blue"
test_b:
    <<: *red
    want: "red"
test_c:
    <<: *red
    <<: *green
    want: "red"
test_d:
    <<: *green
    <<: *red
    want: "green"
test_e:
    <<: *green
    <<: *green
    want: "green"
test_f:
    got: "blue"
    <<: *green
    <<: *red
    want: "blue"
test_g:
    <<: *green
    got: "blue"
    <<: *red
    want: "blue"
test_h:
    <<: *green
    <<: *red
    got: "blue"
    want: "blue"`)

	for _, tt := range []string{
		"test_a",
		"test_b",
		"test_c",
		"test_d",
		"test_e",
		"test_f",
		"test_g",
		"test_h",
	} {
		t.Run(tt, func(t *testing.T) {
			t.Parallel()

			got, err := yamlx.Find(node, tt, "got")
			if err != nil {
				t.Errorf("error finding got: %v", err)
			}
			want, err := yamlx.Find(node, tt, "want")
			if err != nil {
				t.Errorf("error finding want: %v", err)
			}
			if got.Value != want.Value {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
