package model

import (
	"errors"
	"fmt"
	"sync"
)

// ResolvedPos means resolved position. You can resolve a position to get more
// information about it. Objects of this class represent such a resolved
// position, providing various pieces of context information, and some helper
// methods.
//
// Throughout this interface, methods that take an optional depth parameter
// will interpret undefined as this.depth and negative numbers as this.depth +
// value.
type ResolvedPos struct {
	// The position that was resolved.
	Pos  int
	Path []interface{}
	// The number of levels the parent node is from the root. If this
	// position points directly into the root node, it is 0. If it
	// points into a top-level paragraph, 1, and so on.
	Depth int
	// The offset this position has into its parent node.
	ParentOffset int
}

// NewResolvedPos is the constructor of ResolvedPos.
func NewResolvedPos(pos int, path []interface{}, parentOffset int) *ResolvedPos {
	return &ResolvedPos{
		Pos:          pos,
		Path:         path,
		Depth:        len(path)/3 - 1,
		ParentOffset: parentOffset,
	}
}

func (r *ResolvedPos) resolveDepth(val *int) int {
	if val == nil {
		return r.Depth
	}
	if *val < 0 {
		return r.Depth + *val
	}
	return *val
}

// Parent returns the parent node that the position points into. Note that even
// if a position points into a text node, that node is not considered the
// parent—text nodes are ‘flat’ in this model, and have no content.
func (r *ResolvedPos) Parent() *Node {
	return r.Node(r.Depth)
}

// Doc is the root node in which the position was resolved.
func (r *ResolvedPos) Doc() *Node {
	return r.Node(0)
}

// Node returns the ancestor node at the given level. p.node(p.depth) is the
// same as p.parent.
func (r *ResolvedPos) Node(depth ...int) *Node {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	return r.Path[r.resolveDepth(d)*3].(*Node)
}

// Index returns the index into the ancestor at the given level. If this points
// at the 3rd node in the 2nd paragraph on the top level, for example,
// p.index(0) is 1 and p.index(1) is 2.
func (r *ResolvedPos) Index(depth ...int) int {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	return r.Path[r.resolveDepth(d)*3+1].(int)
}

// IndexAfter returns the index pointing after this position into the ancestor
// at the given level.
func (r *ResolvedPos) IndexAfter(depth ...int) int {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	rd := r.resolveDepth(d)
	offset := 0
	if rd == r.Depth && r.TextOffset() == 0 {
		offset = 1
	}
	return r.Index(rd) + offset
}

// Start is the (absolute) position at the start of the node at the given
// level.
func (r *ResolvedPos) Start(depth ...int) int {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	rd := r.resolveDepth(d)
	if rd == 0 {
		return 0
	}
	return r.Path[rd*3-1].(int) + 1
}

// End is the (absolute) position at the end of the node at the given level.
func (r *ResolvedPos) End(depth ...int) int {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	rd := r.resolveDepth(d)
	return r.Start(rd) + r.Node(rd).Content.Size
}

// Before is the (absolute) position directly before the wrapping node at the
// given level, or, when depth is this.depth + 1, the original position.
func (r *ResolvedPos) Before(depth ...int) (int, error) {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	rd := r.resolveDepth(d)
	if rd == 0 {
		return 0, errors.New("There is no position before the top-level node")
	}
	if rd == r.Depth+1 {
		return r.Pos, nil
	}
	return r.Path[rd*3-1].(int), nil
}

// After is the (absolute) position directly after the wrapping node at the
// given level, or the original position when depth is this.depth + 1.
func (r *ResolvedPos) After(depth ...int) (int, error) {
	var d *int
	if len(depth) > 0 {
		d = &depth[0]
	}
	rd := r.resolveDepth(d)
	if rd == 0 {
		return 0, errors.New("There is no position after the top-level node")
	}
	if rd == r.Depth+1 {
		return r.Pos, nil
	}
	return r.Path[rd*3-1].(int) + r.Path[rd*3].(*Node).NodeSize(), nil
}

// TextOffset returns, when this position points into a text node, the distance
// between the position and the start of the text node. Will be zero for
// positions that point between nodes.
func (r *ResolvedPos) TextOffset() int {
	return r.Pos - r.Path[len(r.Path)-1].(int)
}

// NodeAfter gets the node directly after the position, if any. If the position
// points into a text node, only the part of that node after the position is
// returned.
func (r *ResolvedPos) NodeAfter() (*Node, error) {
	parent := r.Parent()
	index := r.Index(r.Depth)
	if index == parent.ChildCount() {
		return nil, nil
	}
	dOff := r.Pos - r.Path[len(r.Path)-1].(int)
	child, err := parent.Child(index)
	if err != nil {
		return nil, err
	}
	if dOff > 0 {
		return child.Cut(dOff), nil
	}
	return child, nil
}

// NodeBefore gets the node directly before the position, if any. If the
// position points into a text node, only the part of that node before the
// position is returned.
func (r *ResolvedPos) NodeBefore() (*Node, error) {
	index := r.Index(r.Depth)
	dOff := r.Pos - r.Path[len(r.Path)-1].(int)
	if dOff > 0 {
		child, err := r.Parent().Child(index)
		if err != nil {
			return nil, err
		}
		return child.Cut(0, dOff), nil
	}
	if index == 0 {
		return nil, nil
	}
	child, err := r.Parent().Child(index - 1)
	if err != nil {
		return nil, err
	}
	return child, nil
}

// Marks gets the marks at this position, factoring in the surrounding marks'
// inclusive property. If the position is at the start of a non-empty node, the
// marks of the node after it (if any) are returned.
func (r *ResolvedPos) Marks() []*Mark {
	parent := r.Parent()
	index := r.Index()

	// In an empty parent, return the empty array
	if parent.Content.Size == 0 {
		return NoMarks
	}

	// When inside a text node, just return the text node's marks
	if r.TextOffset() > 0 {
		child, err := parent.Child(index)
		if err != nil {
			panic(err)
		}
		return child.Marks
	}

	main := parent.MaybeChild(index - 1)
	other := parent.MaybeChild(index)
	// If the after flag is true or there is no node before, make the node
	// after this position the main reference.
	if main == nil {
		main, other = other, main
	}

	// Use all marks in the main node, except those that have inclusive set to
	// false and are not present in the other node.
	marks := main.Marks
	for _, m := range main.Marks {
		if (m.Type.Spec.Inclusive != nil && !*m.Type.Spec.Inclusive) &&
			(other == nil || !m.IsInSet(other.Marks)) {
			marks = m.RemoveFromSet(marks)
		}
	}
	return marks
}

// SharedDepth is the depth up to which this position and the given
// (non-resolved) position share the same parent nodes.
func (r *ResolvedPos) SharedDepth(pos int) int {
	for depth := r.Depth; depth > 0; depth-- {
		if r.Start(depth) <= pos && r.End(depth) >= pos {
			return depth
		}
	}
	return 0
}

func resolvePos(doc *Node, pos int) (*ResolvedPos, error) {
	if !(pos >= 0 && pos <= doc.Content.Size) {
		return nil, fmt.Errorf("Position %d out of range", pos)
	}
	path := []interface{}{}
	start := 0
	parentOffset := pos
	node := doc
	for {
		index, offset, err := node.Content.findIndex(parentOffset)
		if err != nil {
			return nil, err
		}
		rem := parentOffset - offset
		path = append(path, node, index, start+offset)
		if rem == 0 {
			break
		}
		node, err = node.Child(index)
		if err != nil {
			return nil, err
		}
		if node.IsText() {
			break
		}
		parentOffset = rem - 1
		start += offset + 1
	}
	return NewResolvedPos(pos, path, parentOffset), nil
}

func resolvePosCached(doc *Node, pos int) (*ResolvedPos, error) {
	resolveCacheMutex.Lock()
	defer resolveCacheMutex.Unlock()
	for _, entry := range resolveCache {
		if entry.doc == doc && entry.pos.Pos == pos {
			return entry.pos, nil
		}
	}
	result, err := resolvePos(doc, pos)
	if err != nil {
		return nil, err
	}
	resolveCache[resolveCachePos] = resolveEntry{doc, result}
	resolveCachePos = (resolveCachePos + 1) % len(resolveCache)
	return result, nil
}

type resolveEntry struct {
	doc *Node
	pos *ResolvedPos
}

var (
	resolveCacheMutex sync.Mutex
	resolveCache      = make([]resolveEntry, 12)
	resolveCachePos   = 0
)
