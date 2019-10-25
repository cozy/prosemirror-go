package model

import (
	"fmt"
	"reflect"
)

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

// Test whether two nodes represent the same piece of document.
func (n *Node) Eq(other *Node) bool {
	if n == other {
		return true
	}
	return n.SameMarkup(other) && n.Content.Eq(other.Content)
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

// Create a new node with the same markup as this node, containing
// the given content (or empty, if no content is given).
func (n *Node) Copy(content ...*Fragment) *Node {
	c := EmptyFragment
	if len(content) > 0 {
		c = content[0]
	}
	return NewNode(n.Type, n.Attrs, c, n.Marks)
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

// Create a copy of this node with only the content between the given
// positions. If `to` is not given, it defaults to the end of the node.
func (n *Node) Cut(from int, to ...int) *Node {
	if n.IsText() {
		t := len(*n.Text)
		if len(to) > 0 {
			t = to[0]
		}
		if from == 0 && t == len(*n.Text) {
			return n
		}
		return n.WithText((*n.Text)[from:t])
	}
	if len(to) == 0 {
		return n.Copy(n.Content.Cut(from))
	}
	t := to[0]
	if from == 0 && t == n.Content.Size {
		return n
	}
	return n.Copy(n.Content.Cut(from, t))
}

// True when this is a leaf node.
func (n *Node) IsLeaf() bool {
	return n.Type.IsLeaf()
}

// Return a string representation of this node for debugging purposes.
func (n *Node) String() string {
	name := n.Type.Name
	if n.IsText() {
		name = fmt.Sprintf("%q", *n.Text)
	} else if n.Content.Size > 0 {
		name += fmt.Sprintf("(%s)", n.Content.toStringInner())
	}
	return wrapMarks(n.Marks, name)
}

func NewTextNode(typ *NodeType, attrs map[string]interface{}, text string, marks []*Mark) *Node {
	return &Node{Type: typ, Attrs: attrs, Text: &text, Content: EmptyFragment, Marks: marks}
}

// True when this is a text node.
func (n *Node) IsText() bool {
	return n.Text != nil
}

func (n *Node) WithText(text string) *Node {
	if text == *n.Text {
		return n
	}
	return NewTextNode(n.Type, n.Attrs, text, n.Marks)
}

// TODO

func wrapMarks(marks []*Mark, str string) string {
	for i := len(marks) - 1; i >= 0; i-- {
		str = fmt.Sprintf("%s(%s)", marks[i].Type.Name, str)
	}
	return str
}
