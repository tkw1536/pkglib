package yamlx

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Find attempts to find the yaml node with the given path inside a yaml tree.
// A path is a set of strings, and assumed to be keys inside yaml mapping nodes.
// The empty path returns (a possibly de-referenced version of) node.
//
// The returned node is guaranteed to never be a document or alias node, these are resolved automatically.
// As such find should not be used on untrusted input.
//
// If the node does not exist, it returns a nil yaml.Node and an error.
// A nil node being passed in is considered an error.
func Find(node *yaml.Node, path ...string) (*yaml.Node, error) {
	switch {
	// no tree provided => can't find the path
	case node == nil:
		return nil, ErrNodeIsNil

		// document node => directly iterate into the children
	case node.Kind == yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, ChildNotFound("")
		}

		errs := make([]error, len(node.Content))
		for i, child := range node.Content {
			node, err := Find(child, path...)
			if err == nil {
				return node, nil
			}
			errs[i] = err
		}
		return nil, errors.Join(errs...)

	// if we have an alias, find the alias instead
	case node.Kind == yaml.AliasNode:
		// resolve the alias
		node := resolveAlias(node)
		if node == nil {
			return node, ErrNodeIsNil
		}

		// and find in the resolved node
		return Find(node, path...)

		// zero length path => return the current child!
	case len(path) == 0:
		return node, nil

	case node.Kind == yaml.ScalarNode:
		// cannot find a child with a name inside a scalar node
		return node, ErrUnexpectedScalar
	case node.Kind == yaml.SequenceNode:
		// cannot find a child within a sequence node
		return node, ErrUnexpectedSequence
	}

	// find the child with the given index
	child, err := Child(node, path[0])
	if err != nil {
		return nil, err
	}

	// and replace the node
	result, err := Find(child, path[1:]...)
	if err != nil {
		return nil, ChildError{Child: path[0], Err: err}
	}
	return result, nil
}

// Child finds the child node with the given name.
// If it does not exist, it returns an error.
func Child(node *yaml.Node, name string) (*yaml.Node, error) {
	// we must have a mapping node
	// or we cannot find the element with the given key
	if node == nil {
		return nil, ErrNodeIsNil
	}
	if node.Kind != yaml.MappingNode {
		return nil, ErrExpectedMapping
	}

	// do a first pass to find all directly used elements
	var merged []*yaml.Node
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Kind != yaml.ScalarNode {
			return nil, MappingExpectedScalar(i)
		}

		// saw a merge tag => record it and keep going
		if key.Tag == "!!merge" {
			merged = append(merged, node.Content[i+1])
			continue
		}

		// check if we have a string key with the appropriate name
		if key.Tag != "!!str" {
			continue
		}

		if key.Value == name {
			return node.Content[i+1], nil
		}
	}

	// follow merges recursively
	for _, node := range merged {
		child, err := Child(resolveAlias(node), name)
		if err == nil {
			return child, nil
		}
	}

	// nothing found (not even in the merges)
	return nil, ChildNotFound(name)
}

// resolveAlias resolves an alias recursively.
// If the node is not an alias, it returns it unchanged.
func resolveAlias(node *yaml.Node) *yaml.Node {
	// NOTE(twiesing): The yaml spec forbids circular references.
	// Section 7.1 says "The alias refers to the most recent preceding node having the same anchor".
	// This means this loop should terminate if the node comes from valid yaml.
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node
}

type findError string

func (fe findError) Error() string {
	return string(fe)
}

// Common errors for finding a node
var (
	ErrNodeIsNil          error = findError("node is nil")
	ErrUnexpectedScalar   error = findError("unexpected scalar node")
	ErrUnexpectedSequence error = findError("unexpected sequence node")
	ErrExpectedMapping    error = findError("expected mapping node")
)

// ChildNotFound indicates that the given child was not found
type ChildNotFound string

func (cnf ChildNotFound) Error() string {
	return fmt.Sprintf("child not found: %q", string(cnf))
}

// ChildError indicates an error that has occurred inside a specific child node
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
