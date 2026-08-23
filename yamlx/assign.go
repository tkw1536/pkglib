//spellchecker:words yamlx
package yamlx

//spellchecker:words gopkg yaml pkglib errorsx
import (
	"errors"

	"gopkg.in/yaml.v3"
)

// Assign sets the key to the value "value" in the given node with kind [yaml.MappingNode].
// If the key does not exist, a new key: value pair is appended.
//
// Iff mappingNode is not of kind [yaml.MappingNode], a non-nil error is returned.
func Assign(mappingNode *yaml.Node, key string, value yaml.Node) error {
	if mappingNode == nil || mappingNode.Kind != yaml.MappingNode {
		return ErrExpectedMapping
	}

	valueNode, err := Child(mappingNode, key)
	if err != nil {
		_, ok := errors.AsType[ChildNotFoundError](err)
		if !ok {
			panic("never reached: mappingNode is of kind yaml.MappingNode")
		}

		// create a new key and value node
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		}
		valueNode = &yaml.Node{}

		mappingNode.Content = append(mappingNode.Content, keyNode, valueNode)
	}

	Apply(valueNode, value)
	return nil
}
