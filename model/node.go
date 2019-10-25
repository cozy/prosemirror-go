package model

import "reflect"

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
	// The type of node that this is.
	Type *NodeType
	// An object mapping attribute names to values. The kind of attributes
	// allowed and required are determined by the node type.
	Attrs map[string]interface{}
	// A container holding the node's children.
	Content *Fragment
	// For text nodes, this contains the node's text content.
	Text *string
	// The marks (things like whether it is emphasized or part of a link)
	// applied to this node.
	Marks []*Mark
}

func NewNode(typ *NodeType, attrs map[string]interface{}, content *Fragment, marks []*Mark) *Node {
	return &Node{Type: typ, Attrs: attrs, Content: content, Marks: marks}
}

// The size of this node, as defined by the integer-based indexing scheme. For
// text nodes, this is the amount of characters. For other leaf nodes, it is
// one. For non-leaf nodes, it is the size of the content plus two (the start
// and end token).
func (n *Node) NodeSize() int {
	if n.IsText() {
		return len(*n.Text)
	}
	if n.IsLeaf() {
		return 1
	}
	return 2 + n.Content.Size
}

// The number of children that the node has.
func (n *Node) ChildCount() int {
	return n.Content.ChildCount()
}

// Get the child node at the given index. Raises an error when the index is out
// of range.
func (n *Node) Child(index int) (*Node, error) {
	return n.Content.Child(index)
}

// Get the child node at the given index, if it exists.
func (n *Node) MaybeChild(index int) *Node {
	return n.Content.MaybeChild(index)
}

// Compare the markup (type, attributes, and marks) of this node to those of
// another. Returns true if both have the same markup.
func (n *Node) SameMarkup(other *Node) bool {
	return n.HasMarkup(other.Type, other.Attrs, other.Marks)
}

// Check whether this node's markup correspond to the given type, attributes,
// and marks.
// :: (NodeType, ?Object, ?[Mark]) → bool
func (n *Node) HasMarkup(typ *NodeType, args ...interface{}) bool {
	if n.Type != typ {
		return false
	}
	var attrs map[string]interface{}
	if len(args) > 0 {
		attrs, _ = args[0].(map[string]interface{})
	}
	// TODO type.defaultAttrs
	if !reflect.DeepEqual(n.Attrs, attrs) {
		return false
	}
	marks := NoMarks
	if len(args) > 1 {
		marks, _ = args[1].([]*Mark)
	}
	return SameMarkSet(n.Marks, marks)
}

// Create a copy of this node, with the given set of marks instead of the
// node's own marks.
func (n *Node) Mark(marks []*Mark) *Node {
	if sameMarks(n.Marks, marks) {
		return n
	}
	if n.IsText() {
		return NewTextNode(n.Type, n.Attrs, *n.Text, marks)
	}
	return NewNode(n.Type, n.Attrs, n.Content, marks)
}

// True when this is a leaf node.
func (n *Node) IsLeaf() bool {
	return n.Type.IsLeaf()
}

func NewTextNode(typ *NodeType, attrs map[string]interface{}, text string, marks []*Mark) *Node {
	return &Node{Type: typ, Attrs: attrs, Text: &text, Content: EmptyFragment, Marks: marks}
}

// True when this is a text node.
func (n *Node) IsText() bool {
	return n.Text != nil
}

// TODO
