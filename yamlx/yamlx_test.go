//spellchecker:words yamlx
package yamlx_test

//spellchecker:words testing github pkglib yamlx gopkg yaml
import (
	"fmt"
	"testing"

	"github.com/tkw1536/pkglib/yamlx"
	"gopkg.in/yaml.v3"
)

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
	node, err := yamlx.Marshal(value)
	if err != nil {
		panic(err)
	}

	// and print it out
	fmt.Println(mustMarshal(nil, node))

	// Output: count: 2
	// numbers:
	//     "42": the answer
	//     "69": nice
}

func mustMarshal(t testing.TB, node *yaml.Node) string {
	result, err := yaml.Marshal(node)
	if err != nil {
		msg := fmt.Sprintf("unable to marshal: %v", err)
		if t != nil {
			t.Fatalf(msg)
		}
		panic(msg)
	}
	return string(result)
}

func mustUnmarshal(t testing.TB, source string) *yaml.Node {
	var node yaml.Node
	err := yaml.Unmarshal([]byte(source), &node)
	if err != nil {
		msg := fmt.Sprintf("unable to unmarshal: %v", err)
		if t != nil {
			t.Fatalf(msg)
		}
		panic(msg)
	}
	return &node
}
