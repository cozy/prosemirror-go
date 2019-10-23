package model

import "fmt"

// A fragment represents a node's collection of child nodes.
//
// Like nodes, fragments are persistent data structures, and you should not
// mutate them or their content. Rather, you create new instances whenever
// needed. The API tries to make this easy.
type Fragment struct {
	Content []*Node
	Size    int
}

// The number of child nodes in this fragment.
func (f *Fragment) ChildCount() int {
	return len(f.Content)
}

// Get the child node at the given index. Raise an error when the index is out
// of range.
func (f *Fragment) Child(index int) *Node {
	if index >= len(f.Content) {
		panic(fmt.Errorf("Index %d out of range for %v", index, f))
	}
	return f.Content[index]
}

// Find the first position at which this fragment and another fragment differ,
// or `null` if they are the same.
func (f *Fragment) FindDiffStart(other *Fragment, pos int) *int {
	return findDiffStart(f, other, pos)
}

// Find the first position, searching from the end, at which this fragment and
// the given fragment differ, or `null` if they are the same. Since this
// position will not be the same in both nodes, an object with two separate
// positions is returned.
func (f *Fragment) FindDiffEnd(other *Fragment, posA, posB int) *DiffEnd {
	return findDiffEnd(f, other, posA, posB)
}

// TODO
