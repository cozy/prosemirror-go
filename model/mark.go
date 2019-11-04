package model

import (
	"fmt"
	"reflect"
	"sort"
)

// Mark is a piece of information that can be attached to a node, such as it
// being emphasized, in code font, or a link. It has a type and optionally a
// set of attributes that provide further information (such as the target of
// the link). Marks are created through a Schema, which controls which types
// exist and which attributes they have.
type Mark struct {
	Type  *MarkType
	Attrs map[string]interface{}
}

// NewMark is the constructor for Mark.
func NewMark(typ *MarkType, attrs map[string]interface{}) *Mark {
	return &Mark{Type: typ, Attrs: attrs}
}

// AddToSet , when given a set of marks, creates a new set which contains this
// one as well, in the right position. If this mark is already in the set, the
// set itself is returned. If any marks that are set to be exclusive with this
// mark are present, those are replaced by this one.
func (m *Mark) AddToSet(set []*Mark) []*Mark {
	var cpy []*Mark
	placed := false
	for i, other := range set {
		if m.Eq(other) {
			return set
		}
		if m.Type.Excludes(other.Type) {
			if cpy == nil {
				cpy = make([]*Mark, i)
				copy(cpy, set[:i])
			}
		} else if other.Type.Excludes(m.Type) {
			return set
		} else {
			if !placed && other.Type.Rank > m.Type.Rank {
				if cpy == nil {
					cpy = make([]*Mark, i)
					copy(cpy, set[:i])
				}
				cpy = append(cpy, m)
				placed = true
			}
			if cpy != nil {
				cpy = append(cpy, set[i])
			}
		}
	}
	if cpy == nil {
		cpy = make([]*Mark, len(set))
		copy(cpy, set)
	}
	if !placed {
		cpy = append(cpy, m)
	}
	return cpy
}

// RemoveFromSet removes this mark from the given set, returning a new set. If
// this mark is not in the set, the set itself is returned.
func (m *Mark) RemoveFromSet(set []*Mark) []*Mark {
	for i, other := range set {
		if m.Eq(other) {
			cpy := make([]*Mark, len(set)-1)
			copy(cpy[:i], set[:i])
			copy(cpy[i:], set[i+1:])
			return cpy
		}
	}
	return set
}

// IsInSet tests whether this mark is in the given set of marks.
func (m *Mark) IsInSet(set []*Mark) bool {
	for _, other := range set {
		if m.Eq(other) {
			return true
		}
	}
	return false
}

// Eq tests whether this mark has the same type and attributes as another mark.
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
	return reflect.DeepEqual(m.Attrs, other.Attrs)
}

// ToJSON converts this mark to a JSON-serializeable representation.
func (m *Mark) ToJSON() map[string]interface{} {
	obj := map[string]interface{}{"type": m.Type.Name}
	if len(m.Attrs) > 0 {
		obj["attrs"] = m.Attrs
	}
	return obj
}

// MarkFromJSON deserializes a mark from its JSON representation.
func MarkFromJSON(schema *Schema, raw map[string]interface{}) (*Mark, error) {
	t, _ := raw["type"].(string)
	typ, ok := schema.Marks[t]
	if !ok {
		return nil, fmt.Errorf("There is no mark %s in this schema", raw["type"])
	}
	attrs, _ := raw["attrs"].(map[string]interface{})
	return typ.Create(attrs), nil
}

func sameMarks(a, b []*Mark) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SameMarkSet tests whether two sets of marks are identical.
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

// MarkSetFrom creates a properly sorted mark set from null, a single mark, or
// an unsorted array of marks.
func MarkSetFrom(marks ...interface{}) []*Mark {
	if len(marks) == 0 {
		return NoMarks
	}
	if mark, ok := marks[0].(*Mark); ok {
		return []*Mark{mark}
	}
	if marks, ok := marks[0].([]*Mark); ok {
		set := make([]*Mark, len(marks))
		copy(set, marks)
		sort.Slice(set, func(i, j int) bool {
			return set[i].Type.Rank < set[j].Type.Rank
		})
		return set
	}
	panic(fmt.Errorf("Unexpected marks for MarkSetFrom: %#v", marks))
}

// NoMarks is the empty set of marks (none in JS)
var NoMarks = []*Mark{}
