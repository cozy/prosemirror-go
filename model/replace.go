package model

import "fmt"

// A slice represents a piece cut out of a larger document. It stores not only
// a fragment, but also the depth up to which nodes on both side are ‘open’
// (cut through).
type Slice struct {
	// Fragment The slice's content.
	Content *Fragment
	// The open depth at the start.
	OpenStart int
	// number The open depth at the end.
	OpenEnd int
}

// Create a slice. When specifying a non-zero open depth, you must make sure
// that there are nodes of at least that depth at the appropriate side of the
// fragment—i.e. if the fragment is an empty paragraph node, openStart and
// openEnd can't be greater than 1.
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

// The size this slice would add when inserted into a document.
func (s *Slice) Size() int {
	return s.Content.Size - s.OpenStart - s.OpenEnd
}

// Tests whether this slice is equal to another slice.
func (s *Slice) Eq(other *Slice) bool {
	return s.Content.Eq(other.Content) && s.OpenStart == other.OpenStart && s.OpenEnd == other.OpenEnd
}

func (s *Slice) String() string {
	return fmt.Sprintf("%s(%d,%d)", s.Content.String(), s.OpenStart, s.OpenEnd)
}

var EmptySlice = NewSlice(EmptyFragment, 0, 0)
