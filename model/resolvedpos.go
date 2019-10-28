package model

import (
	"errors"
	"fmt"
)

// You can resolve a position to get more information about it. Objects of this
// class represent such a resolved position, providing various pieces of
// context information, and some helper methods.
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

// The parent node that the position points into. Note that even if a position
// points into a text node, that node is not considered the parent—text nodes
// are ‘flat’ in this model, and have no content.
func (r *ResolvedPos) Parent() *Node {
	return r.Node(&r.Depth)
}

// The root node in which the position was resolved.
func (r *ResolvedPos) Doc() *Node {
	zero := 0
	return r.Node(&zero)
}

// The ancestor node at the given level. `p.node(p.depth)` is the same as
// `p.parent`.
func (r *ResolvedPos) Node(depth *int) *Node {
	return r.Path[r.resolveDepth(depth)*3].(*Node)
}

// The (absolute) position at the start of the node at the given level.
func (r *ResolvedPos) Start(depth *int) int {
	d := r.resolveDepth(depth)
	if d == 0 {
		return 0
	}
	return r.Path[d*3-1].(int) + 1
}

// The (absolute) position at the end of the node at the given level.
func (r *ResolvedPos) End(depth *int) int {
	d := r.resolveDepth(depth)
	return r.Start(&d) + r.Node(&d).Content.Size
}

// The (absolute) position directly before the wrapping node at the given
// level, or, when depth is this.depth + 1, the original position.
func (r *ResolvedPos) Before(depth *int) (int, error) {
	d := r.resolveDepth(depth)
	if d == 0 {
		return 0, errors.New("There is no position before the top-level node") // TODO RangeError
	}
	if d == r.Depth+1 {
		return r.Pos, nil
	}
	return r.Path[d*3-1].(int), nil
}

// The (absolute) position directly after the wrapping node at the given level,
// or the original position when depth is this.depth + 1.
func (r *ResolvedPos) After(depth *int) (int, error) {
	d := r.resolveDepth(depth)
	if d == 0 {
		return 0, errors.New("There is no position after the top-level node") // TODO RangeError
	}
	if d == r.Depth+1 {
		return r.Pos, nil
	}
	return r.Path[d*3-1].(int) + r.Path[d*3].(*Node).NodeSize(), nil
}

// The depth up to which this position and the given (non-resolved) position
// share the same parent nodes.
func (r *ResolvedPos) SharedDepth(pos int) int {
	for depth := r.Depth; depth > 0; depth-- {
		if r.Start(&depth) <= pos && r.End(&depth) >= pos {
			return depth
		}
	}
	return 0
}

func resolvePos(doc *Node, pos int) (*ResolvedPos, error) {
	if !(pos >= 0 && pos <= doc.Content.Size) {
		return nil, fmt.Errorf("Position %d out of range", pos) // TODO RangeError
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
	// TODO add cache
	return resolvePos(doc, pos)
}

// TODO
