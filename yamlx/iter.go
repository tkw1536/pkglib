//spellchecker:words yamlx
package yamlx

//spellchecker:words iter slices gopkg yaml
import (
	"iter"
	"slices"

	"gopkg.in/yaml.v3"
)

// Iterate iterates over all paths in node.
//
// Calling Find(node, path.Path) == path.Node is guaranteed for all paths.
// See also [Find].
func Iterate(node *yaml.Node) iter.Seq[Path] {
	return func(yield func(Path) bool) {
		iterPaths(yield, node, nil, nil)
	}
}

// Path represents a path inside a given struct.
type Path struct {
	Path []string
	Node *yaml.Node
}

func (path Path) HasChildren() bool {
	node := resolveAlias(path.Node)
	return node.Kind == yaml.MappingNode
}

// the return value indicates if the caller should continue.
func iterPaths(yield func(Path) bool, node *yaml.Node, path []string, merge_keys map[string]struct{}) bool {
	// resolve the alias
	node = resolveAlias(node)
	if node == nil {
		return false
	}

	// if we got a document node, DO NOT send the node itself.
	// and directly iterate on the children
	if node.Kind == yaml.DocumentNode {
		for _, doc := range node.Content {
			if !iterPaths(yield, doc, path, nil) {
				return false
			}
		}

		return true
	}

	// send the node itself (unless we did the merge)
	if merge_keys == nil {
		if !yield(Path{Path: path, Node: node}) {
			return false
		}
	}

	// iterate over each child
	if node.Kind == yaml.MappingNode {
		var merged []*yaml.Node

		// record the nodes we saw before the merge
		if merge_keys == nil {
			merge_keys = make(map[string]struct{}, len(node.Content)/2)
		}

		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			if key.Kind != yaml.ScalarNode {
				continue
			}
			// record the merge tag
			if key.Tag == "!!merge" {
				merged = append(merged, node.Content[i+1])
				continue
			}
			if key.Tag != "!!str" {
				continue
			}

			// we already saw the child in the parent
			if _, saw := merge_keys[key.Value]; saw {
				continue
			}

			// recursively iterate the children
			path := append(slices.Clone(path), key.Value)
			if !iterPaths(yield, value, path, nil) {
				return false
			}
			merge_keys[key.Value] = struct{}{}
		}

		// iterate through the merged children
		for _, node := range merged {
			if !iterPaths(yield, node, path, merge_keys) {
				return false
			}
		}
	}

	return true
}
