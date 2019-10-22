package model

// A mark is a piece of information that can be attached to a node, such as it
// being emphasized, in code font, or a link. It has a type and optionally a
// set of attributes that provide further information (such as the target of
// the link). Marks are created through a Schema, which controls which types
// exist and which attributes they have.
type Mark struct {
	Type  *MarkType
	Attrs map[string]string
}

// Given a set of marks, create a new set which contains this one as well, in
// the right position. If this mark is already in the set, the set itself is
// returned. If any marks that are set to be exclusive with this mark are
// present, those are replaced by this one.
func (m *Mark) AddToSet(set []*Mark) []*Mark {
	var cpy []*Mark
	placed := false
	for i, other := range set {
		if m.Eq(other) {
			return set
		}
		if m.Type.Excludes(other.Type) {
			if cpy == nil {
				cpy := make([]*Mark, i)
				copy(cpy, set[:i])
			}
		} else if other.Type.Excludes(m.Type) {
			return set
		} else {
			if !placed && other.Type.Rank > m.Type.Rank {
				if cpy == nil {
					cpy := make([]*Mark, i)
					copy(cpy, set[:i])
				}
				cpy = append(cpy, m)
				placed = true
			}
			if cpy != nil {
				cpy = append(cpy, other)
			}
		}
	}
	if cpy == nil {
		cpy := make([]*Mark, len(set))
		copy(cpy, set)
	}
	if !placed {
		cpy = append(cpy, m)
	}
	return cpy
}

// Remove this mark from the given set, returning a new set. If this mark is
// not in the set, the set itself is returned.
func (m *Mark) RemoveFromSet(set []*Mark) []*Mark {
	for i, other := range set {
		if m.Eq(other) {
			cpy := make([]*Mark, len(set)-1)
			copy(cpy[:i], set[:i])
			copy(cpy[i:], set[i+1:])
		}
	}
	return set
}

// Test whether this mark is in the given set of marks.
func (m *Mark) IsInSet(set []*Mark) bool {
	for _, other := range set {
		if m.Eq(other) {
			return true
		}
	}
	return false
}

// Test whether this mark has the same type and attributes as another mark.
func (m *Mark) Eq(other *Mark) bool {
	if m == other {
		return true
	}
	if m.Type != other.Type {
		return false
	}
	if len(m.Attrs) != len(other.Attrs) {
		return false
	}
	// TODO use reflect.DeepEqual?
	for k, v := range m.Attrs {
		if other.Attrs[k] != v {
			return false
		}
	}
	return true
}

// TODO Marshal and Unmarshal JSON

// Test whether two sets of marks are identical.
func SameMarkSet(a, b []*Mark) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Eq(b[i]) {
			return false
		}
	}
	return true
}

// Create a properly sorted mark set from null, a single mark, or an unsorted
// array of marks.
func MarkSetFrom(marks []*Mark) []*Mark {
	if len(marks) == 0 {
		return none
	}
	if len(marks) == 1 {
		return marks
	}
	set := make([]*Mark, len(marks))
	copy(set, marks)
	return set
}

// The empty set of marks.
// TODO export it? with which name?
var none = []*Mark{}
