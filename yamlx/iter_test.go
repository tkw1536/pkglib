//spellchecker:words yamlx
package yamlx_test

//spellchecker:words reflect testing github pkglib yamlx
import (
	"reflect"
	"testing"

	"github.com/tkw1536/pkglib/yamlx"
)

//spellchecker:words bref

func TestIteratePaths(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		Name      string
		Source    string
		wantPaths [][]string
	}{
		{
			"null",
			`null`,
			[][]string{
				nil,
			},
		},
		{
			"single mapping",
			`
a: {}
b: {}
c: {}`,
			[][]string{
				nil,
				{"a"},
				{"b"},
				{"c"},
			},
		},
		{
			"nested mapping",
			`
a: {}
b:
    b1: ""
    b2: ""
c: {}`,
			[][]string{
				nil,
				{"a"},
				{"b"},
				{"b", "b1"},
				{"b", "b2"},
				{"c"},
			},
		},
		{
			"nested mapping with aliases",
			`
a: {}
b: &bref
    b1: ""
    b2: ""
c: {}
d: *bref`,
			[][]string{
				nil,
				{"a"},
				{"b"},
				{"b", "b1"},
				{"b", "b2"},
				{"c"},
				{"d"},
				{"d", "b1"},
				{"d", "b2"},
			},
		},
		{
			"nested mapping with merges",
			`
vehicle: &vehicle
    is_vehicle: true
car: &car
    <<: *vehicle
    is_car: true
bike:
    <<: *car
    is_car: false
`,
			[][]string{
				nil,
				{"vehicle"},
				{"vehicle", "is_vehicle"},
				{"car"},
				{"car", "is_car"},
				{"car", "is_vehicle"},
				{"bike"},
				{"bike", "is_car"},
				{"bike", "is_vehicle"},
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			node := mustUnmarshal(t, tt.Source)

			// iterate over the node and record the paths
			gotPaths := make([][]string, 0, len(tt.wantPaths))
			for datum := range yamlx.Iterate(node) {

				// check that Find(node, path.Path...) == path.Node holds
				wantNode := datum.Node
				gotNode, err := yamlx.Find(node, datum.Path...)
				if err != nil {
					t.Errorf("path %v: gotNode returned error %v", datum.Path, err)
				}

				if !reflect.DeepEqual(gotNode, wantNode) {
					t.Errorf("Find(node, %v) != node", datum.Path)
				}

				// record the found path
				gotPaths = append(gotPaths, datum.Path)
			}

			// check that the paths we got are identical
			if !reflect.DeepEqual(gotPaths, tt.wantPaths) {
				t.Errorf("IteratePaths returned paths = %#v, but want = %#v", gotPaths, tt.wantPaths)
			}
		})
	}
}
