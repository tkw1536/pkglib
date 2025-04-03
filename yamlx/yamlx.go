// Package yamlx provides extended YAML parsing functionalities.
// It supports both references and merges.
//
//spellchecker:words yamlx
package yamlx

//spellchecker:words gopkg yaml
import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Marshal marshals a value into a new yaml node.
func Marshal(value any) (*yaml.Node, error) {
	node := new(yaml.Node)
	err := node.Encode(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encode node: %w", err)
	}
	return node, nil
}
