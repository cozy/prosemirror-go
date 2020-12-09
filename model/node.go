package model

import (
	"errors"
	"fmt"
	"reflect"
	"unicode/utf16"
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
		return len(asCodeUnits(*n.Text))
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

// ForEach calls fn for every child node, passing the node, its offset into this
// parent node, and its index.
func (n *Node) ForEach(fn func(node *Node, offset, index int)) {
	n.Content.ForEach(fn)
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
		units := asCodeUnits(*n.Text)
		return fromCodeUnits(units[from:to])
	}
	return n.Content.textBetween(from, to, args...)
}

// UnitCodeAt returns the UTF-16 unit code at the given position. It is a
// function that does not exist in the original prosemirror in JS, as it
// is only useful in Go to emulate the behavior of strings in JavaScript.
// In Go, the strings are encoded as UTF-8 by default, but in JavaScript,
// they can be seen as an array of UTF-16 unit codes. In particular, it means
// that a single unicode character with a value > 0xffff are a surrogate pair
// of two UTF-16 unit codes in JavaScript.
func (n *Node) UnitCodeAt(pos int) uint16 {
	if n.IsText() {
		return asCodeUnits(*n.Text)[pos]
	}
	return 0
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
		return false
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
		units := asCodeUnits(*n.Text)
		t := len(units)
		if len(to) > 0 {
			t = to[0]
		}
		if from == 0 && t == len(units) {
			return n
		}
		return n.WithText(fromCodeUnits(units[from:t]))
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
func (n *Node) Slice(from int, args ...interface{}) (*Slice, error) {
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
		return EmptySlice, nil
	}

	resFrom, err := n.Resolve(from)
	if err != nil {
		return nil, err
	}
	resTo, err := n.Resolve(to)
	if err != nil {
		return nil, err
	}
	depth := 0
	if !includeParents {
		depth = resFrom.SharedDepth(to)
	}
	start := resFrom.Start(depth)
	node := resFrom.Node(depth)
	content := node.Content.Cut(resFrom.Pos-start, resTo.Pos-start)
	return NewSlice(content, resFrom.Depth-depth, resTo.Depth-depth), nil
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

// ContentMatchAt gets the content match in this node at the given index.
func (n *Node) ContentMatchAt(index int) (*ContentMatch, error) {
	match := n.Type.ContentMatch.MatchFragment(n.Content, 0, index)
	if match == nil {
		return nil, errors.New("Called contentMatchAt on a node with invalid content")
	}
	return match, nil
}

// CanReplace tests whether replacing the range between from and to (by child
// index) with the given replacement fragment (which defaults to the empty
// fragment) would leave the node's content valid. You can optionally pass
// start and end indices into the replacement fragment.
//
// :: (number, number, ?Fragment, ?number, ?number) → bool
func (n *Node) CanReplace(from, to int, args ...interface{}) bool {
	replacement := EmptyFragment
	if len(args) > 0 {
		replacement = args[0].(*Fragment)
	}
	start := 0
	if len(args) > 1 {
		start = args[1].(int)
	}
	var end int
	if len(args) > 2 {
		end = args[2].(int)
	} else {
		end = replacement.ChildCount()
	}
	match, err := n.ContentMatchAt(from)
	if err != nil {
		return false
	}
	one := match.MatchFragment(replacement, start, end)
	var two *ContentMatch
	if one != nil {
		two = one.MatchFragment(n.Content, to)
	}
	if two == nil || !two.ValidEnd {
		return false
	}
	for i := start; i < end; i++ {
		child, err := replacement.Child(i)
		if err == nil && !n.Type.AllowsMarks(child.Marks) {
			return false
		}
	}
	return true
}

// ToJSON converts this node to a JSON-serializeable representation.
func (n *Node) ToJSON() map[string]interface{} {
	obj := map[string]interface{}{"type": n.Type.Name}
	if len(n.Attrs) > 0 {
		obj["attrs"] = n.Attrs
	}
	if n.Content.Size > 0 {
		obj["content"] = n.Content.ToJSON()
	}
	if len(n.Marks) > 0 {
		var marks []interface{}
		for _, m := range n.Marks {
			marks = append(marks, m.ToJSON())
		}
		obj["marks"] = marks
	}
	if n.IsText() {
		obj["text"] = *n.Text
	}
	return obj
}

// NodeFromJSON deserializes a node from its JSON representation.
func NodeFromJSON(schema *Schema, raw map[string]interface{}) (*Node, error) {
	var marks []*Mark
	if data, ok := raw["marks"]; ok {
		items, ok := data.([]interface{})
		if !ok {
			return nil, errors.New("Invalid mark data for Node.fromJSON")
		}
		for _, item := range items {
			obj, _ := item.(map[string]interface{})
			m, err := MarkFromJSON(schema, obj)
			if err != nil {
				return nil, err
			}
			marks = append(marks, m)
		}
	}
	if raw["type"] == "text" {
		text, ok := raw["text"].(string)
		if !ok {
			return nil, errors.New("Invalid text node in JSON")
		}
		return schema.Text(text, marks), nil
	}
	content, err := FragmentFromJSON(schema, raw["content"])
	if err != nil {
		return nil, err
	}
	nodeType, _ := raw["type"].(string)
	typ, err := schema.NodeType(nodeType)
	if err != nil {
		return nil, err
	}
	attrs, _ := raw["attrs"].(map[string]interface{})
	return typ.Create(attrs, content, marks)
}

// NewTextNode is a constructor for text Node.
func NewTextNode(typ *NodeType, attrs map[string]interface{}, text string, marks []*Mark) *Node {
	return &Node{Type: typ, Attrs: attrs, Text: &text, Content: EmptyFragment, Marks: marks}
}

// IsText returns true when this is a text node.
func (n *Node) IsText() bool {
	return n.Text != nil
}

// WithText returns a new text node with the given string.
func (n *Node) WithText(text string) *Node {
	if text == *n.Text {
		return n
	}
	return NewTextNode(n.Type, n.Attrs, text, n.Marks)
}

func wrapMarks(marks []*Mark, str string) string {
	for i := len(marks) - 1; i >= 0; i-- {
		str = fmt.Sprintf("%s(%s)", marks[i].Type.Name, str)
	}
	return str
}

func asCodeUnits(text string) []uint16 {
	return utf16.Encode([]rune(text))
}

func fromCodeUnits(units []uint16) string {
	return string(utf16.Decode(units))
}
