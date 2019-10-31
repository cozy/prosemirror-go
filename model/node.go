package model

import (
	"fmt"
	"reflect"
)

// Node class represents a node in the tree that makes up a ProseMirror
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

// NewNode is the constructor of Node.
func NewNode(typ *NodeType, attrs map[string]interface{}, content *Fragment, marks []*Mark) *Node {
	return &Node{Type: typ, Attrs: attrs, Content: content, Marks: marks}
}

// NodeSize returns the size of this node, as defined by the integer-based
// indexing scheme. For text nodes, this is the amount of characters. For other
// leaf nodes, it is one. For non-leaf nodes, it is the size of the content
// plus two (the start and end token).
func (n *Node) NodeSize() int {
	if n.IsText() {
		return len(*n.Text)
	}
	if n.IsLeaf() {
		return 1
	}
	return 2 + n.Content.Size
}

// ChildCount returns the number of children that the node has.
func (n *Node) ChildCount() int {
	return n.Content.ChildCount()
}

// Child gets the child node at the given index. Raises an error when the index
// is out of range.
func (n *Node) Child(index int) (*Node, error) {
	return n.Content.Child(index)
}

// MaybeChild gets the child node at the given index, if it exists.
func (n *Node) MaybeChild(index int) *Node {
	if index < 0 {
		return nil
	}
	return n.Content.MaybeChild(index)
}

// NodesBetween invokes a callback for all descendant nodes recursively between
// the given two positions that are relative to start of this node's content.
// The callback is invoked with the node, its parent-relative position, its
// parent node, and its child index. When the callback returns false for a
// given node, that node's children will not be recursed over. The last
// parameter can be used to specify a starting position to count from.
func (n *Node) NodesBetween(from, to int, fn NBCallback, startPos ...int) {
	s := 0
	if len(startPos) > 0 {
		s = startPos[0]
	}
	n.Content.NodesBetween(from, to, fn, s, n)
}

// TextContent concatenates all the text nodes found in this fragment and its
// children.
func (n *Node) TextContent() string {
	if n.IsText() {
		return *n.Text
	}
	return n.TextBetween(0, n.Content.Size, "")
}

// TextBetween gets all text between positions from and to. When blockSeparator
// is given, it will be inserted whenever a new block node is started. When
// leafText is given, it'll be inserted for every non-text leaf node
// encountered.
func (n *Node) TextBetween(from, to int, args ...string) string {
	if n.IsText() {
		return (*n.Text)[from:to]
	}
	return n.Content.textBetween(from, to, args...)
}

// FirstChild returns this node's first child, or null if there are no
// children.
func (n *Node) FirstChild() *Node {
	return n.Content.FirstChild()
}

// LastChild returns this node's last child, or null if there are no children.
func (n *Node) LastChild() *Node {
	return n.Content.LastChild()
}

// Eq tests whether two nodes represent the same piece of document.
func (n *Node) Eq(other *Node) bool {
	if n == other {
		return true
	}
	return n.SameMarkup(other) && n.Content.Eq(other.Content)
}

// SameMarkup compares the markup (type, attributes, and marks) of this node to
// those of another. Returns true if both have the same markup.
func (n *Node) SameMarkup(other *Node) bool {
	return n.HasMarkup(other.Type, other.Attrs, other.Marks)
}

// HasMarkup checks whether this node's markup correspond to the given type,
// attributes, and marks.
//
// :: (NodeType, ?Object, ?[Mark]) → bool
func (n *Node) HasMarkup(typ *NodeType, args ...interface{}) bool {
	if n.Type != typ {
		return false
	}
	var attrs map[string]interface{}
	if len(args) > 0 {
		attrs, _ = args[0].(map[string]interface{})
	} else {
		attrs = typ.DefaultAttrs
	}
	if !reflect.DeepEqual(n.Attrs, attrs) {
		// TODO fix this bug
		if _, ok := n.Attrs["nodeType"]; ok {
			return false
		}
		nt, ok := attrs["nodeType"]
		if !ok {
			return false
		}
		if n.Attrs == nil {
			n.Attrs = map[string]interface{}{}
		}
		n.Attrs["nodeType"] = nt
		if !reflect.DeepEqual(n.Attrs, attrs) {
			return false
		}
		// TODO return false
	}
	marks := NoMarks
	if len(args) > 1 {
		marks, _ = args[1].([]*Mark)
	}
	return SameMarkSet(n.Marks, marks)
}

// Copy creates a new node with the same markup as this node, containing the
// given content (or empty, if no content is given).
func (n *Node) Copy(content ...*Fragment) *Node {
	c := EmptyFragment
	if len(content) > 0 {
		c = content[0]
	}
	return NewNode(n.Type, n.Attrs, c, n.Marks)
}

// Mark creates a copy of this node, with the given set of marks instead of the
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

// Cut creates a copy of this node with only the content between the given
// positions. If to is not given, it defaults to the end of the node.
func (n *Node) Cut(from int, to ...int) *Node {
	if n.IsText() {
		t := len(*n.Text)
		if len(to) > 0 {
			t = to[0]
		}
		if from == 0 && t == len(*n.Text) {
			return n
		}
		return n.withText((*n.Text)[from:t])
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

// Slice cuts out the part of the document between the given positions, and
// return it as a Slice object.
func (n *Node) Slice(from int, args ...interface{}) *Slice {
	to := n.Content.Size
	if len(args) > 0 {
		if t, ok := args[0].(int); ok {
			to = t
		}
	}
	includeParents := false
	if len(args) > 1 {
		includeParents, _ = args[1].(bool)
	}

	if from == to {
		return EmptySlice
	}

	resFrom, err := n.Resolve(from)
	if err != nil {
		panic(err)
	}
	resTo, err := n.Resolve(to)
	if err != nil {
		panic(err)
	}
	depth := 0
	if !includeParents {
		depth = resFrom.SharedDepth(to)
	}
	start := resFrom.Start(depth)
	node := resFrom.Node(depth)
	content := node.Content.Cut(resFrom.Pos-start, resTo.Pos-start)
	return NewSlice(content, resFrom.Depth-depth, resTo.Depth-depth)
}

// Replace the part of the document between the given positions with the given
// slice. The slice must 'fit', meaning its open sides must be able to connect
// to the surrounding content, and its content nodes must be valid children for
// the node they are placed into. If any of this is violated, an error of type
// ReplaceError is thrown.
func (n *Node) Replace(from, to int, slice *Slice) (*Node, error) {
	f, err := n.Resolve(from)
	if err != nil {
		return nil, err
	}
	t, err := n.Resolve(to)
	if err != nil {
		return nil, err
	}
	return replace(f, t, slice)
}

// Resolve the given position in the document, returning an object with
// information about its context.
func (n *Node) Resolve(pos int) (*ResolvedPos, error) {
	return resolvePosCached(n, pos)
}

func (n *Node) resolveNoCache(pos int) (*ResolvedPos, error) {
	return resolvePos(n, pos)
}

// NodeAt finds the node directly after the given position.
func (n *Node) NodeAt(pos int) *Node {
	node := n
	for {
		index, offset, err := node.Content.findIndex(pos)
		if err != nil {
			panic(err)
		}
		node = node.MaybeChild(index)
		if node == nil {
			return nil
		}
		if offset == pos || node.IsText() {
			return node
		}
		pos -= offset + 1
	}
}

// IsBlock returns true when this is a block (non-inline node)
func (n *Node) IsBlock() bool {
	return n.Type.IsBlock()
}

// IsInline returns true when this is an inline node (a text node or a node
// that can appear among text).
func (n *Node) IsInline() bool {
	return n.Type.IsInline()
}

// IsLeaf returns true when this is a leaf node.
func (n *Node) IsLeaf() bool {
	return n.Type.IsLeaf()
}

// String returns a string representation of this node for debugging purposes.
func (n *Node) String() string {
	if n.Type.Spec.ToDebugString != nil {
		return n.Type.Spec.ToDebugString(n)
	}
	name := n.Type.Name
	if n.IsText() {
		name = fmt.Sprintf("%q", *n.Text)
	} else if n.Content.Size > 0 {
		name += fmt.Sprintf("(%s)", n.Content.toStringInner())
	}
	return wrapMarks(n.Marks, name)
}

// NewTextNode is a constructor for text Node.
func NewTextNode(typ *NodeType, attrs map[string]interface{}, text string, marks []*Mark) *Node {
	return &Node{Type: typ, Attrs: attrs, Text: &text, Content: EmptyFragment, Marks: marks}
}

// IsText returns true when this is a text node.
func (n *Node) IsText() bool {
	return n.Text != nil
}

func (n *Node) withText(text string) *Node {
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
