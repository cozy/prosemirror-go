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

// TODO
