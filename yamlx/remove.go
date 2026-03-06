//spellchecker:words yamlx
package yamlx

//spellchecker:words gopkg yaml
import "gopkg.in/yaml.v3"

var (
	ErrZeroLengthPath error = findError("path is empty")
)

// Remove removes the node at the given path from a YAML document.
// A path is a set of strings, and assumed to be keys inside yaml mapping nodes.
// The path must be of non-zero length.
//
// Remove first finds the parent node using the path prefix, then removes the child
// with the name specified by the last element of the path. The parent must be a
// mapping node, otherwise an error is returned.
//
// If any part of the path does not exist, or if the parent is not a mapping node,
// an error is returned. A nil node being passed in is considered an error.
func Remove(node *yaml.Node, path ...string) error {
	if len(path) == 0 {
		return ErrZeroLengthPath
	}

	// Find the parent node
	node, err := Find(node, path[:len(path)-1]...)
	if err != nil {
		return err
	}
	if node == nil {
		return ErrNodeIsNil
	}
	if node.Kind != yaml.MappingNode {
		return ErrExpectedMapping
	}

	// Find the child as the index of the parent
	// and then remove that index.
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value != path[len(path)-1] {
			continue
		}
		node.Content = append(node.Content[:i], node.Content[i+2:]...)
		return nil
	}
	return ChildNotFoundError(path[len(path)-1])
}
