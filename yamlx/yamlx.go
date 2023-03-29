package yamlx

import (
	"fmt"

	"github.com/tkw1536/pkglib/iterator"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// Marshal marshals a value into a new yaml node
func Marshal(value any) (*yaml.Node, error) {
	node := new(yaml.Node)
	err := node.Encode(value)
	return node, err
}

type FindError string

func (fe FindError) Error() string {
	return string(fe)
}

var (
	NodeIsNil          = FindError("node is nil")
	UnexpectedScalar   = FindError("unexpected scalar node")
	UnexpectedSequence = FindError("unexpected sequence node")
	MappingExpected    = FindError("expected mapping node")
)

type ChildNotFound string

func (cnf ChildNotFound) Error() string {
	return fmt.Sprintf("child not found: %q", string(cnf))
}

type ChildError struct {
	Child string
	Err   error
}

func (pe ChildError) Error() string {
	return fmt.Sprintf("Node %q: %s", pe.Child, pe.Err)
}

func (pe ChildError) Unwrap() error {
	return pe.Err
}

type MappingExpectedScalar int

func (mes MappingExpectedScalar) Error() string {
	return fmt.Sprintf("expected scalar node in content with index %d", int(mes))
}

// Path represents a path inside a given struct
type Path struct {
	Path []string
	Node *yaml.Node
}

func (path Path) HasChildren() bool {
	node := path.Node
	for node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node.Kind == yaml.MappingNode
}

// Transplant transplants all nodes found inside donor onto node.
//
// donor and node should be of the same shape, meaning for every path where
//
//	Find(donor, path...)
//
// does not return an error should also not return an error in node.
func Transplant(node, donor *yaml.Node) error {
	it := IteratePaths(donor)
	defer it.Close()

	for it.Next() {
		path := it.Datum()

		if path.HasChildren() {
			continue
		}

		if err := Replace(node, *path.Node, path.Path...); err != nil {
			return err
		}
	}

	return nil
}

// Iterate iterates over all paths in node.
func IteratePaths(node *yaml.Node) iterator.Iterator[Path] {
	return iterator.New(func(g iterator.Generator[Path]) {
		defer g.Return()
		iterpaths(g, node, nil)
	})
}

// iterpaths generates all paths in the given node.
// It returns if the user requested cancellation.
func iterpaths(g iterator.Generator[Path], node *yaml.Node, path []string) bool {
	if node == nil {
		return false
	}

	// send the node itself
	if g.Yield(Path{Path: path, Node: node}) {
		return true
	}

	// resolve the alias
	// TODO: Implement the limit here
	for node.Kind == yaml.AliasNode {
		node = node.Alias
	}

	if node.Kind != yaml.MappingNode {
		return false
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]
		if key.Kind != yaml.ScalarNode {
			continue
		}
		if key.Tag != "!!str" {
			continue
		}

		// iterate over the children
		path := append(slices.Clone(path), key.Value)
		if iterpaths(g, value, path) {
			return true
		}
	}

	return true

}

// Replace replaces the node found by Find(node, path...) with replacement.
// If the original node is an anchor, it will not be replaced.
// If the original node is not an anchor, it will be replaced
func Replace(node *yaml.Node, replacement yaml.Node, path ...string) error {
	// TODO: Test me
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
	// TODO: Test me
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

// Find attempts to find the yaml node with the given path inside a yaml tree.
// Find should not be used on untrustued input; it follows anchors by default.
// If the node does not exist, it returns nil.
func Find(node *yaml.Node, path ...string) (*yaml.Node, error) {
	switch {
	// no tree provided => can't find the path
	case node == nil:
		return nil, NodeIsNil

		// document node => directly iterate into the children
	case node.Kind == yaml.DocumentNode:
		lastErr := error(ChildNotFound(""))
		for _, child := range node.Content {
			node, err := Find(child, path...)
			if err == nil {
				return node, nil
			}
			lastErr = err
		}
		return nil, lastErr
	// if we have an alias, find the alias instead
	case node.Kind == yaml.AliasNode:
		return Find(node.Alias, path...)

		// zero length path => return the current child!
	case len(path) == 0:
		return node, nil

	case node.Kind == yaml.ScalarNode:
		// cannot find a child with a name inside a scalar node
		return node, UnexpectedScalar
	case node.Kind == yaml.SequenceNode:
		// cannot find a child within a sequence node
		return node, UnexpectedSequence
	}

	// find the child with the given index
	index, err := Child(node, path[0])
	if err != nil {
		return nil, err
	}

	// and replace the node
	result, err := Find(node.Content[index], path[1:]...)
	if err != nil {
		return nil, ChildError{Child: path[0], Err: err}
	}
	return result, nil
}

// Child finds the child node with the given name.
// Find is safe to be used on untrusted input; it does not follow anchors.
func Child(node *yaml.Node, name string) (index int, err error) {
	if node == nil {
		return -1, NodeIsNil
	}
	if node.Kind != yaml.MappingNode {
		return -1, MappingExpected
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Kind != yaml.ScalarNode {
			return -1, MappingExpectedScalar(i)
		}
		if key.Tag != "!!str" {
			continue
		}
		if key.Value == name {
			return i + 1, nil
		}
	}
	return -1, ChildNotFound(name)
}
