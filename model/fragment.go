package model

import (
	"errors"
	"fmt"
)

// Fragment represents a node's collection of child nodes.
//
// Like nodes, fragments are persistent data structures, and you should not
// mutate them or their content. Rather, you create new instances whenever
// needed. The API tries to make this easy.
type Fragment struct {
	Content []*Node
	// Size is the total of the size of its content nodes.
	Size int
}

// NewFragment is the constructor for Fragment.
func NewFragment(content []*Node, size ...int) *Fragment {
	fragment := Fragment{Content: content, Size: 0}
	if len(size) == 0 {
		for _, node := range content {
			fragment.Size += node.NodeSize()
		}
	} else {
		fragment.Size = size[0]
	}
	return &fragment
}

// NBCallback is a type of the function used for NodesBetween. The arguments
// are:
// - the current node
// - the current position
// - the parent node
// - the index of the current node in the list of its parent children.
// If the callback returns false, it will prevent NodesBetween to descend
// into this node.
type NBCallback func(*Node, int, *Node, int) bool

// NodesBetween invokes a callback for all descendant nodes between the given
// two positions (relative to start of this fragment). Doesn't descend into a
// node when the callback returns false.
func (f *Fragment) NodesBetween(from, to int, fn NBCallback, nodeStart int, parent *Node) *int {
	pos := 0
	for i, child := range f.Content {
		if pos >= to {
			break
		}
		end := pos + child.NodeSize()
		if end > from && fn(child, nodeStart+pos, parent, i) && child.Content.Size > 0 {
			start := pos + 1
			f := 0
			if x := from - start; x > 0 {
				f = x
			}
			t := child.Content.Size
			if x := to - start; x < t {
				t = x
			}
			child.NodesBetween(f, t, fn, nodeStart+start)
		}
		pos = end
	}
	return nil
}

func (f *Fragment) textBetween(from, to int, args ...string) string {
	blockSeparator := ""
	if len(args) > 0 {
		blockSeparator = args[0]
	}
	leafText := ""
	if len(args) > 1 {
		leafText = args[1]
	}
	text := ""
	separated := true
	f.NodesBetween(from, to, func(node *Node, pos int, _ *Node, _ int) bool {
		if node.IsText() {
			max := from
			if pos > max {
				max = pos
			}
			start := max - pos
			stop := to - pos
			if stop > len(*node.Text) {
				stop = len(*node.Text)
			}
			text += (*node.Text)[start:stop]
			separated = blockSeparator != ""
		} else if node.IsLeaf() && leafText != "" {
			text += leafText
			separated = blockSeparator != ""
		} else if !separated && node.IsBlock() {
			text += blockSeparator
			separated = true
		}
		return true
	}, 0, nil)
	return text
}

// Append creates a new fragment containing the combined content of this
// fragment and the other.
func (f *Fragment) Append(other *Fragment) *Fragment {
	if other.Size == 0 {
		return f
	}
	if f.Size == 0 {
		return other
	}
	last := f.LastChild()
	first := other.FirstChild()
	content := make([]*Node, len(f.Content))
	copy(content, f.Content)
	i := 0
	if last.IsText() && last.SameMarkup(first) {
		content[len(content)-1] = last.withText(*last.Text + *first.Text)
		i = 1
	}
	content = append(content, other.Content[i:]...)
	return NewFragment(content, f.Size+other.Size)
}

// Cut out the sub-fragment between the two given positions.
func (f *Fragment) Cut(from int, to ...int) *Fragment {
	t := f.Size
	if len(to) > 0 {
		t = to[0]
	}
	if from == 0 && t == f.Size {
		return f
	}
	result := []*Node{}
	size := 0
	if t > from {
		pos := 0
		for _, child := range f.Content {
			if pos >= t {
				break
			}
			end := pos + child.NodeSize()
			if end > from {
				if pos < from || end > t {
					var start, stop int
					if child.IsText() {
						if x := from - pos; x >= 0 {
							start = x
						}
						stop = len(*child.Text)
						if x := t - pos; x < stop {
							stop = x
						}
					} else {
						if x := from - pos - 1; x >= 0 {
							start = x
						}
						stop = child.Content.Size
						if x := t - pos - 1; x < stop {
							stop = x
						}
					}
					child = child.Cut(start, stop)
				}
				result = append(result, child)
				size += child.NodeSize()
			}
			pos = end
		}
	}
	return NewFragment(result, size)
}

// ReplaceChild creates a new fragment in which the node at the given index is
// replaced by the given node.
func (f *Fragment) ReplaceChild(index int, node *Node) *Fragment {
	current := f.Content[index]
	if current == node {
		return f
	}
	cpy := make([]*Node, len(f.Content))
	copy(cpy, f.Content)
	size := f.Size + node.NodeSize() - current.NodeSize()
	cpy[index] = node
	return NewFragment(cpy, size)
}

// Eq compares this fragment to another one.
func (f *Fragment) Eq(other *Fragment) bool {
	if len(f.Content) != len(other.Content) {
		return false
	}
	for i, node := range f.Content {
		if !node.Eq(other.Content[i]) {
			return false
		}
	}
	return true
}

// FirstChild returns the first child of the fragment, or null if it is empty.
func (f *Fragment) FirstChild() *Node {
	if len(f.Content) == 0 {
		return nil
	}
	return f.Content[0]
}

// LastChild returns the last child of the fragment, or null if it is empty.
func (f *Fragment) LastChild() *Node {
	if len(f.Content) == 0 {
		return nil
	}
	return f.Content[len(f.Content)-1]
}

// ChildCount returns the number of child nodes in this fragment.
func (f *Fragment) ChildCount() int {
	return len(f.Content)
}

// Child gets the child node at the given index. Raise an error when the index
// is out of range.
func (f *Fragment) Child(index int) (*Node, error) {
	if index >= len(f.Content) {
		return nil, fmt.Errorf("Index %d out of range for %v", index, f)
	}
	return f.Content[index], nil
}

// MaybeChild gets the child node at the given index, if it exists.
func (f *Fragment) MaybeChild(index int) *Node {
	if index >= len(f.Content) {
		return nil
	}
	return f.Content[index]
}

// FindDiffStart finds the first position at which this fragment and another
// fragment differ, or null if they are the same.
func (f *Fragment) FindDiffStart(other *Fragment, pos ...int) *int {
	p := 0
	if len(pos) > 0 {
		p = pos[0]
	}
	return findDiffStart(f, other, p)
}

// FindDiffEnd finds the first position, searching from the end, at which this
// fragment and the given fragment differ, or `null` if they are the same.
// Since this position will not be the same in both nodes, an object with two
// separate positions is returned.
func (f *Fragment) FindDiffEnd(other *Fragment, pos ...int) *DiffEnd {
	posA := f.Size
	posB := other.Size
	if len(pos) > 0 {
		posA = pos[0]
	}
	if len(pos) > 1 {
		posB = pos[1]
	}
	return findDiffEnd(f, other, posA, posB)
}

// findIndex finds the index and inner offset corresponding to a given relative
// position in this fragment.
func (f *Fragment) findIndex(pos int, round ...int) (index int, offset int, err error) {
	if pos == 0 {
		return 0, pos, nil
	}
	if pos == f.Size {
		return len(f.Content), pos, nil
	}
	if pos > f.Size || pos < 0 {
		return 0, 0, fmt.Errorf("Position %d outside of fragment (%v)", pos, f)
	}
	r := -1
	if len(round) > 0 {
		r = round[0]
	}
	curPos := 0
	for i, cur := range f.Content {
		end := curPos + cur.NodeSize()
		if end >= pos {
			if end == pos || r > 0 {
				return i + 1, end, nil
			}
			return i, curPos, nil
		}
		curPos = end
	}
	panic(errors.New("Unexpected state"))
}

// String returns a debugging string that describes this fragment.
func (f *Fragment) String() string {
	return fmt.Sprintf("<%s>", f.toStringInner())
}

func (f *Fragment) toStringInner() string {
	str := ""
	for i, node := range f.Content {
		if i > 0 {
			str += ", "
		}
		str += node.String()
	}
	return str
}

// TODO

// FragmentFromArray builds a fragment from an array of nodes. Ensures that
// adjacent text nodes with the same marks are joined together.
func FragmentFromArray(array []*Node) *Fragment {
	if len(array) == 0 {
		return EmptyFragment
	}
	var joined []*Node
	size := 0
	for i, node := range array {
		size += node.NodeSize()
		if i > 0 && node.IsText() && array[i-1].SameMarkup(node) {
			if len(joined) == 0 {
				joined = array[0:i]
			}
			was := joined[len(joined)-1].Text
			joined[len(joined)-1] = node.withText(*was + *node.Text)
		} else if len(joined) > 0 {
			joined = append(joined, node)
		}
	}
	if len(joined) == 0 {
		joined = array
	}
	return NewFragment(joined, size)
}

// FragmentFrom creates a fragment from something that can be interpreted as a
// set of nodes. For null, it returns the empty fragment. For a fragment, the
// fragment itself. For a node or array of nodes, a fragment containing those
// nodes.
func FragmentFrom(nodes interface{}) (*Fragment, error) {
	if nodes == nil {
		return EmptyFragment, nil
	}
	switch nodes := nodes.(type) {
	case *Fragment:
		return nodes, nil
	case []*Node:
		return FragmentFromArray(nodes), nil
	case []interface{}:
		array := make([]*Node, len(nodes))
		for i := range nodes {
			n, ok := nodes[i].(*Node)
			if !ok {
				return nil, fmt.Errorf("Invalid type for FragmentFrom: %#v", nodes)
			}
			array[i] = n
		}
		return FragmentFromArray(array), nil
	case *Node:
		return NewFragment([]*Node{nodes}, nodes.NodeSize()), nil
	}
	return nil, fmt.Errorf("Can not convert %v to a Fragment", nodes)
}

// EmptyFragment is an empty fragment.
var EmptyFragment = &Fragment{Content: nil, Size: 0}
