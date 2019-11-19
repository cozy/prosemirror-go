package model

import (
	"errors"
	"fmt"
)

// ReplaceError is the error type raised by Node.replace when given an invalid
// replacement.
type ReplaceError struct {
	Message string
}

// NewReplaceError is the constructor for ReplaceError.
func NewReplaceError(message string, args ...interface{}) *ReplaceError {
	return &ReplaceError{Message: fmt.Sprintf(message, args...)}
}

// Error returns the error message.
func (e *ReplaceError) Error() string {
	return e.Message
}

// Slice represents a piece cut out of a larger document. It stores not only a
// fragment, but also the depth up to which nodes on both side are ‘open’ (cut
// through).
type Slice struct {
	// Fragment The slice's content.
	Content *Fragment
	// The open depth at the start.
	OpenStart int
	// number The open depth at the end.
	OpenEnd int
}

// NewSlice creates a slice. When specifying a non-zero open depth, you must
// make sure that there are nodes of at least that depth at the appropriate
// side of the fragment—i.e. if the fragment is an empty paragraph node,
// openStart and openEnd can't be greater than 1.
//
// It is not necessary for the content of open nodes to conform to the schema's
// content constraints, though it should be a valid start/end/middle for such a
// node, depending on which sides are open.
func NewSlice(content *Fragment, openStart, openEnd int) *Slice {
	return &Slice{
		Content:   content,
		OpenStart: openStart,
		OpenEnd:   openEnd,
	}
}

// Size returns the size this slice would add when inserted into a document.
func (s *Slice) Size() int {
	return s.Content.Size - s.OpenStart - s.OpenEnd
}

// InsertAt inserts a fragment at the given position.
func (s *Slice) InsertAt(pos int, fragment *Fragment) *Slice {
	content, _ := insertInto(s.Content, pos+s.OpenStart, fragment, nil)
	if content == nil {
		return nil
	}
	return NewSlice(content, s.OpenStart, s.OpenEnd)
}

// RemoveBetween removes content between the two given positions.
func (s *Slice) RemoveBetween(from, to int) (*Slice, error) {
	removed, err := removeRange(s.Content, from+s.OpenStart, to+s.OpenStart)
	if err != nil {
		return nil, err
	}
	return NewSlice(removed, s.OpenStart, s.OpenEnd), nil
}

// Eq tests whether this slice is equal to another slice.
func (s *Slice) Eq(other *Slice) bool {
	return s.Content.Eq(other.Content) && s.OpenStart == other.OpenStart && s.OpenEnd == other.OpenEnd
}

// String returns a string representation of this slice.
func (s *Slice) String() string {
	return fmt.Sprintf("%s(%d,%d)", s.Content.String(), s.OpenStart, s.OpenEnd)
}

// ToJSON converts a slice to a JSON-serializable representation.
func (s *Slice) ToJSON() interface{} {
	if s.Content.Size == 0 {
		return nil
	}
	obj := map[string]interface{}{
		"content": s.Content.ToJSON(),
	}
	if s.OpenStart > 0 {
		obj["openStart"] = s.OpenStart
	}
	if s.OpenEnd > 0 {
		obj["openEnd"] = s.OpenEnd
	}
	return obj
}

// SliceFromJSON deserializes a slice from its JSON representation.
func SliceFromJSON(schema *Schema, obj interface{}) (*Slice, error) {
	if obj == nil {
		return EmptySlice, nil
	}
	data, ok := obj.(map[string]interface{})
	if !ok {
		return EmptySlice, nil
	}
	var openStart, openEnd int
	switch o := data["openStart"].(type) {
	case int:
		openStart = o
	case float64:
		openStart = int(o)
	}
	switch o := data["openEnd"].(type) {
	case int:
		openEnd = o
	case float64:
		openEnd = int(o)
	}
	fragment, err := FragmentFromJSON(schema, data["content"])
	if err != nil {
		return nil, err
	}
	return NewSlice(fragment, openStart, openEnd), nil
}

func removeRange(content *Fragment, from, to int) (*Fragment, error) {
	index, offset, err := content.findIndex(from)
	if err != nil {
		return nil, err
	}
	child := content.MaybeChild(index)
	indexTo, offsetTo, err := content.findIndex(to)
	if err != nil {
		return nil, err
	}
	if offset == from || child.IsText() {
		if offsetTo != to {
			child, err := content.Child(indexTo)
			if err != nil {
				return nil, err
			}
			if !child.IsText() {
				return nil, errors.New("Removing non-flat range")
			}
		}
		return content.Cut(0, from).Append(content.Cut(to)), nil
	}
	if index != indexTo {
		return nil, errors.New("Removing non-flat range")
	}
	removed, err := removeRange(child.Content, from-offset-1, to-offset-1)
	if err != nil {
		return nil, err
	}
	return content.ReplaceChild(index, child.Copy(removed)), nil
}

func insertInto(content *Fragment, dist int, insert *Fragment, parent *Node) (*Fragment, error) {
	index, offset, err := content.findIndex(dist)
	if err != nil {
		return nil, err
	}
	child := content.MaybeChild(index)
	if offset == dist || child.IsText() {
		if parent != nil && !parent.CanReplace(index, index, insert) {
			return nil, nil
		}
		return content.Cut(0, dist).Append(insert).Append(content.Cut(dist)), nil
	}
	inner, err := insertInto(child.Content, dist-offset-1, insert, nil)
	if err != nil {
		return nil, err
	}
	if inner == nil {
		return nil, nil
	}
	return content.ReplaceChild(index, child.Copy(inner)), nil
}

// EmptySlice is an empty slice.
var EmptySlice = NewSlice(EmptyFragment, 0, 0)

func replace(from, to *ResolvedPos, slice *Slice) (*Node, error) {
	if slice.OpenStart > from.Depth {
		return nil, NewReplaceError("Inserted content deeper than insertion position")
	}
	if from.Depth-slice.OpenStart != to.Depth-slice.OpenEnd {
		return nil, NewReplaceError("Inconsistent open depths")
	}
	return replaceOuter(from, to, slice, 0)
}

func replaceOuter(from, to *ResolvedPos, slice *Slice, depth int) (*Node, error) {
	index := from.Index(depth)
	node := from.Node(depth)
	if index == to.Index(depth) && depth < from.Depth-slice.OpenStart {
		inner, err := replaceOuter(from, to, slice, depth+1)
		if err != nil {
			return nil, err
		}
		return node.Copy(node.Content.ReplaceChild(index, inner)), nil
	} else if slice.Content.Size == 0 {
		replaced, err := replaceTwoWay(from, to, depth)
		if err != nil {
			return nil, err
		}
		return replaceClose(node, replaced)
	} else if slice.OpenStart == 0 && slice.OpenEnd == 0 && from.Depth == depth && to.Depth == depth { // Simple, flat case
		parent := from.Parent()
		content := parent.Content.
			Cut(0, from.ParentOffset).
			Append(slice.Content).
			Append(parent.Content.Cut(to.ParentOffset))
		return replaceClose(parent, content)
	}
	start, end, err := prepareSliceForReplace(slice, from)
	if err != nil {
		return nil, err
	}
	replaced, err := replaceThreeWay(from, start, end, to, depth)
	if err != nil {
		return nil, err
	}
	return replaceClose(node, replaced)
}

func checkJoin(main, sub *Node) error {
	if !sub.Type.compatibleContent(main.Type) {
		return NewReplaceError("Cannot join %s onto %s", sub.Type.Name, main.Type.Name)
	}
	return nil
}

func joinable(before, after *ResolvedPos, depth int) (*Node, error) {
	node := before.Node(depth)
	if err := checkJoin(node, after.Node(depth)); err != nil {
		return nil, err
	}
	return node, nil
}

func addNode(child *Node, target []*Node) []*Node {
	last := len(target) - 1
	if last >= 0 && child.IsText() && child.SameMarkup(target[last]) {
		target[last] = child.WithText(*target[last].Text + *child.Text)
	} else {
		target = append(target, child)
	}
	return target
}

func addRange(start, end *ResolvedPos, depth int, target []*Node) ([]*Node, error) {
	r := end
	if r == nil {
		r = start
	}
	node := r.Node(depth)
	endIndex := 0
	if end == nil {
		endIndex = node.ChildCount()
	} else {
		endIndex = end.Index(depth)
	}
	startIndex := 0
	if start != nil {
		startIndex = start.Index(depth)
		if start.Depth > depth {
			startIndex++
		} else if start.TextOffset() != 0 {
			n, err := start.NodeAfter()
			if err != nil {
				return nil, err
			}
			target = addNode(n, target)
			startIndex++
		}
	}
	for i := startIndex; i < endIndex; i++ {
		n, err := node.Child(i)
		if err != nil {
			return nil, err
		}
		target = addNode(n, target)
	}
	if end != nil && end.Depth == depth && end.TextOffset() != 0 {
		n, err := end.NodeBefore()
		if err != nil {
			return nil, err
		}
		target = addNode(n, target)
	}
	return target, nil
}

// replaceClose in Go is close in JS (close is a reserved keyword in go).
func replaceClose(node *Node, content *Fragment) (*Node, error) {
	if !node.Type.ValidContent(content) {
		return nil, NewReplaceError("Invalid content for node %s", node.Type.Name)
	}
	return node.Copy(content), nil
}

func replaceThreeWay(from, start, end, to *ResolvedPos, depth int) (*Fragment, error) {
	var err error
	var openStart *Node
	if from.Depth > depth {
		openStart, err = joinable(from, start, depth+1)
		if err != nil {
			return nil, err
		}
	}
	var openEnd *Node
	if to.Depth > depth {
		openEnd, err = joinable(end, to, depth+1)
		if err != nil {
			return nil, err
		}
	}

	content, err := addRange(nil, from, depth, nil)
	if err != nil {
		return nil, err
	}
	if openStart != nil && openEnd != nil && start.Index(depth) == end.Index(depth) {
		if err = checkJoin(openStart, openEnd); err != nil {
			return nil, err
		}
		replaced, err := replaceThreeWay(from, start, end, to, depth+1)
		if err != nil {
			return nil, err
		}
		closed, err := replaceClose(openStart, replaced)
		if err != nil {
			return nil, err
		}
		content = addNode(closed, content)
	} else {
		if openStart != nil {
			replaced, err := replaceTwoWay(from, start, depth+1)
			if err != nil {
				return nil, err
			}
			closed, err := replaceClose(openStart, replaced)
			if err != nil {
				return nil, err
			}
			content = addNode(closed, content)
		}
		if content, err = addRange(start, end, depth, content); err != nil {
			return nil, err
		}
		if openEnd != nil {
			replaced, err := replaceTwoWay(end, to, depth+1)
			if err != nil {
				return nil, err
			}
			closed, err := replaceClose(openEnd, replaced)
			if err != nil {
				return nil, err
			}
			content = addNode(closed, content)
		}
	}
	if content, err = addRange(to, nil, depth, content); err != nil {
		return nil, err
	}
	return NewFragment(content), nil
}

func replaceTwoWay(from, to *ResolvedPos, depth int) (*Fragment, error) {
	content, err := addRange(nil, from, depth, nil)
	if err != nil {
		return nil, err
	}
	if from.Depth > depth {
		typ, err := joinable(from, to, depth+1)
		if err != nil {
			return nil, err
		}
		replaced, err := replaceTwoWay(from, to, depth+1)
		if err != nil {
			return nil, err
		}
		closed, err := replaceClose(typ, replaced)
		if err != nil {
			return nil, err
		}
		content = addNode(closed, content)
	}
	content, err = addRange(to, nil, depth, content)
	if err != nil {
		return nil, err
	}
	return NewFragment(content), nil
}

func prepareSliceForReplace(slice *Slice, along *ResolvedPos) (*ResolvedPos, *ResolvedPos, error) {
	extra := along.Depth - slice.OpenStart
	parent := along.Node(extra)
	node := parent.Copy(slice.Content)
	for i := extra - 1; i >= 0; i-- {
		fragment, err := FragmentFrom(node)
		if err != nil {
			return nil, nil, err
		}
		node = along.Node(i).Copy(fragment)
	}
	start, err := node.resolveNoCache(slice.OpenStart + extra)
	if err != nil {
		return nil, nil, err
	}
	end, err := node.resolveNoCache(node.Content.Size - slice.OpenEnd - extra)
	if err != nil {
		return nil, nil, err
	}
	return start, end, nil
}
