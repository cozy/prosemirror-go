package model

// This class represents a node in the tree that makes up a ProseMirror
// document. So a document is an instance of Node, with children that are also
// instances of Node.
//
// Nodes are persistent data structures. Instead of changing them, you create
// new ones with the content you want. Old ones keep pointing at the old
// document shape. This is made cheaper by sharing structure between the old
// and new data as much as possible, which a tree shape like this (without back
// pointers) makes easy.
//
// Do not directly mutate the properties of a Node object.
type Node struct {
	// TODO it should probably be an interface
	Content *Fragment
}

// The size of this node, as defined by the integer-based indexing scheme. For
// text nodes, this is the amount of characters. For other leaf nodes, it is
// one. For non-leaf nodes, it is the size of the content plus two (the start
// and end token).
func (n *Node) NodeSize() int {
	return 1 // TODO
}

// Compare the markup (type, attributes, and marks) of this node to those of
// another. Returns true if both have the same markup.
func (n *Node) SameMarkup(other *Node) bool {
	return false // TODO
}

// True when this is a text node.
func (n *Node) IsText() bool {
	return false // TODO
}

func (n *Node) Text() string {
	return "" // TODO
}

// TODO
