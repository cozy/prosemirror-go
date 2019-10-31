package transform

import "fmt"

// Mappable is an interface. There are several things that positions can be
// mapped through. Such objects conform to this interface.
type Mappable interface {
	// Map a position through this object. When given, assoc (should be -1 or
	// 1, defaults to 1) determines with which side the position is associated,
	// which determines in which direction to move when a chunk of content is
	// inserted at the mapped position.
	Map(pos int, assoc ...int) int

	// MapResult maps a position, and returns an object containing additional
	// information about the mapping. The result's deleted field tells you
	// whether the position was deleted (completely enclosed in a replaced
	// range) during the mapping. When content on only one side is deleted, the
	// position itself is only considered deleted when assoc points in the
	// direction of the deleted content.
	MapResult(pos int, assoc ...int) *MapResult
}

// TODO recovery

// MapResult is an object representing a mapped position with extra
// information.
type MapResult struct {
	// The mapped version of the position.
	Pos int
	// Tells you whether the position was deleted, that is, whether the step
	// removed its surroundings from the document.
	Deleted bool
}

// NewMapResult is the constructor for MapResult
func NewMapResult(pos int, deleted ...bool) *MapResult {
	d := false
	if len(deleted) > 0 {
		d = deleted[0]
	}
	return &MapResult{Pos: pos, Deleted: d}
}

// StepMap is a map describing the deletions and insertions made by a step,
// which can be used to find the correspondence between positions in the
// pre-step version of a document and the same position in the post-step
// version.
type StepMap struct {
	Ranges   []int
	Inverted bool
}

// NewStepMap creates a position map. The modifications to the document are
// represented as an array of numbers, in which each group of three represents
// a modified chunk as [start, oldSize, newSize].
func NewStepMap(ranges []int, inverted ...bool) *StepMap {
	inv := false
	if len(inverted) > 0 {
		inv = inverted[0]
	}
	return &StepMap{Ranges: ranges, Inverted: inv}
}

// MapResult is part of the Mappable interface.
func (sm *StepMap) MapResult(pos int, assoc ...int) *MapResult {
	a := 1
	if len(assoc) > 0 {
		a = assoc[0]
	}
	return sm._map(pos, a, false).(*MapResult)
}

// Map is part of the Mappable interface.
func (sm *StepMap) Map(pos int, assoc ...int) int {
	a := 1
	if len(assoc) > 0 {
		a = assoc[0]
	}
	return sm._map(pos, a, true).(int)
}

func (sm *StepMap) _map(pos, assoc int, simple bool) interface{} {
	diff := 0
	oldIndex, newIndex := 1, 2
	if sm.Inverted {
		oldIndex, newIndex = 2, 1
	}
	for i := 0; i < len(sm.Ranges); i += 3 {
		start := sm.Ranges[i]
		if sm.Inverted {
			start -= diff
		}
		if start > pos {
			break
		}
		oldSize := sm.Ranges[i+oldIndex]
		newSize := sm.Ranges[i+newIndex]
		end := start + oldSize
		if pos <= end {
			var side int
			if oldSize == 0 {
				side = assoc
			} else if pos == start {
				side = 1
			} else if pos == end {
				side = -1
			} else {
				side = assoc
			}
			result := start + diff
			if side >= 0 {
				result += newSize
			}
			if simple {
				return result
			}
			deleted := pos != end
			if assoc < 0 {
				deleted = pos != start
			}
			return NewMapResult(result, deleted)
		}
		diff += newSize - oldSize
	}
	if simple {
		return pos + diff
	}
	return NewMapResult(pos + diff)
}

// Invert creates an inverted version of this map. The result can be used to
// map positions in the post-step document to the pre-step document.
func (sm *StepMap) Invert() *StepMap {
	return NewStepMap(sm.Ranges, !sm.Inverted)
}

// String returns a string representation of this StepMap.
func (sm *StepMap) String() string {
	prefix := ""
	if sm.Inverted {
		prefix = "-"
	}
	return fmt.Sprintf("%s%v", prefix, sm.Ranges) // TODO JSON.stringify
}

// EmptyStepMap is an empty StepMap.
var EmptyStepMap = NewStepMap(nil)

var _ Mappable = &StepMap{}
