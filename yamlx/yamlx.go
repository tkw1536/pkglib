// Package yamlx provides extended YAML parsing functionalities.
// It supports both references and merges.
//
//spellchecker:words yamlx
package yamlx

//spellchecker:words gopkg yaml
import (
	"gopkg.in/yaml.v3"
)

// Marshal marshals a value into a new yaml node
func Marshal(value any) (*yaml.Node, error) {
	node := new(yaml.Node)
	err := node.Encode(value)
	return node, err
}
