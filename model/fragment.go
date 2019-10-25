package model

import (
	"fmt"
)

// A fragment represents a node's collection of child nodes.
//
// Like nodes, fragments are persistent data structures, and you should not
// mutate them or their content. Rather, you create new instances whenever
// needed. The API tries to make this easy.
type Fragment struct {
	Content []*Node
	// Size is the total of the size of its content nodes.
	Size int
}

func NewFragment(content []*Node, size ...int) *Fragment {
	fragment := Fragment{Content: content, Size: 0}
	if len(size) == 0 {
		for _, node := range content {
			fragment.Size += node.NodeSize()
		}
	} else {
		fragment.Size = size[0]
	}
	return &fragment
}

// The number of child nodes in this fragment.
func (f *Fragment) ChildCount() int {
	return len(f.Content)
}

// Get the child node at the given index. Raise an error when the index is out
// of range.
func (f *Fragment) Child(index int) (*Node, error) {
	if index >= len(f.Content) {
		return nil, fmt.Errorf("Index %d out of range for %v", index, f)
	}
	return f.Content[index], nil
}

// Get the child node at the given index, if it exists.
func (f *Fragment) MaybeChild(index int) *Node {
	if index >= len(f.Content) {
		return nil
	}
	return f.Content[index]
}

// Find the first position at which this fragment and another fragment differ,
// or `null` if they are the same.
func (f *Fragment) FindDiffStart(other *Fragment, pos ...int) *int {
	p := 0
	if len(pos) > 0 {
		p = pos[0]
	}
	return findDiffStart(f, other, p)
}

// Find the first position, searching from the end, at which this fragment and
// the given fragment differ, or `null` if they are the same. Since this
// position will not be the same in both nodes, an object with two separate
// positions is returned.
func (f *Fragment) FindDiffEnd(other *Fragment, pos ...int) *DiffEnd {
	posA := f.Size
	posB := other.Size
	if len(pos) > 0 {
		posA = pos[0]
	}
	if len(pos) > 1 {
		posB = pos[1]
	}
	return findDiffEnd(f, other, posA, posB)
}

func (f *Fragment) toStringInner() string {
	str := ""
	for i, node := range f.Content {
		if i > 0 {
			str += ", "
		}
		str += node.String()
	}
	return str
}

// TODO

// Build a fragment from an array of nodes. Ensures that adjacent text nodes
// with the same marks are joined together.
func FragmentFromArray(array []*Node) *Fragment {
	if len(array) == 0 {
		return EmptyFragment
	}
	// TODO
	return NewFragment(array)
}

// Create a fragment from something that can be interpreted as a set of nodes.
// For null, it returns the empty fragment. For a fragment, the fragment
// itself. For a node or array of nodes, a fragment containing those nodes.
func FragmentFrom(nodes interface{}) (*Fragment, error) {
	if nodes == nil {
		return EmptyFragment, nil
	}
	switch nodes := nodes.(type) {
	case *Fragment:
		return nodes, nil
	case []*Node:
		return FragmentFromArray(nodes), nil
	case *Node:
		return NewFragment([]*Node{nodes}, nodes.NodeSize()), nil
	}
	return nil, fmt.Errorf("Can not convert %v to a Fragment", nodes)
}

var EmptyFragment = &Fragment{Content: nil, Size: 0}
