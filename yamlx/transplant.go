package yamlx

import "gopkg.in/yaml.v3"

// Transplant transplants all nodes found inside donor onto node.
//
// Unless SkipMissing is true, donor and node should be of the same shape.
// Being of the same shape means every path where
//
//	Find(donor, path...)
//
// does not return an error should also not return an error in node.
func Transplant(node, donor *yaml.Node, SkipMissing bool) error {
	it := Iterate(donor)
	defer it.Close()

	for it.Next() {
		path := it.Datum()

		if path.HasChildren() {
			continue
		}

		err := Replace(node, *path.Node, path.Path...)
		if err != nil && !SkipMissing {
			return err
		}
	}

	return nil
}

// ReplaceWith is like Replace, except that the replacement is first marshalled to yaml.
func ReplaceWith(node *yaml.Node, replacement any, path ...string) error {
	mNode, err := Marshal(replacement)
	if err != nil {
		return err
	}
	return Replace(node, *mNode, path...)
}

// Replace replaces the node found by Find(node, path...) with replacement.
// If the original node is an anchor, it will not be replaced.
// If the original node is not an anchor, it will be replaced
func Replace(node *yaml.Node, replacement yaml.Node, path ...string) error {
	found, err := Find(node, path...)
	if err != nil {
		return err
	}
	Apply(found, replacement)
	return nil
}

// Apply applies a replacement to a node.
//
// If the node is nil, it is not replaced.
// Otherwise, the following fields are copied:
// - Kind
// - Style
// - Tag
// - Value
// - Alias
// - Content
// Note that the original comments are maintained.
func Apply(node *yaml.Node, replacement yaml.Node) {
	if node == nil {
		return
	}
	node.Kind = replacement.Kind
	node.Style = replacement.Style
	node.Tag = replacement.Tag
	node.Value = replacement.Value
	node.Alias = replacement.Alias
	node.Content = replacement.Content
}
