package transform

import (
	"errors"

	"github.com/cozy/prosemirror-go/model"
)

// ReplaceStep replaces a part of the document with a slice of new content.
type ReplaceStep struct {
	From      int
	To        int
	Slice     *model.Slice
	Structure bool
}

// NewReplaceStep is the constructor of ReplaceStep.
//
// The given slice should fit the 'gap' between from and to—the depths must
// line up, and the surrounding nodes must be able to be joined with the open
// sides of the slice. When structure is true, the step will fail if the
// content between from and to is not just a sequence of closing and then
// opening tokens (this is to guard against rebased replace steps overwriting
// something they weren't supposed to).
func NewReplaceStep(from, to int, slice *model.Slice, structure ...bool) *ReplaceStep {
	s := false
	if len(structure) > 0 {
		s = structure[0]
	}
	return &ReplaceStep{From: from, To: to, Slice: slice, Structure: s}
}

// Apply is a method of the Step interface.
func (s *ReplaceStep) Apply(doc *model.Node) StepResult {
	if s.Structure && contentBetween(doc, s.From, s.To) {
		return Fail("Structure replace would overwrite content")
	}
	return FromReplace(doc, s.From, s.To, s.Slice)
}

// GetMap is a method of the Step interface.
func (s *ReplaceStep) GetMap() *StepMap {
	return NewStepMap([]int{s.From, s.To - s.From, s.Slice.Size()})
}

// Invert is a method of the Step interface.
func (s *ReplaceStep) Invert(doc *model.Node) Step {
	slice, err := doc.Slice(s.From, s.To)
	if err != nil {
		panic(err)
	}
	return NewReplaceStep(s.From, s.From+s.Slice.Size(), slice)
}

// Map is a method of the Step interface.
func (s *ReplaceStep) Map(mapping Mappable) Step {
	from := mapping.MapResult(s.From, 1)
	to := mapping.MapResult(s.To, -1)
	if from.Deleted && to.Deleted {
		return nil
	}
	max := from.Pos
	if to.Pos > max {
		max = to.Pos
	}
	return NewReplaceStep(from.Pos, max, s.Slice)
}

// Merge is a method of the Step interface.
func (s *ReplaceStep) Merge(other Step) (Step, bool) {
	repl, ok := other.(*ReplaceStep)
	if !ok || repl.Structure || s.Structure {
		return nil, false
	}
	if s.From+s.Slice.Size() == repl.From && s.Slice.OpenStart == 0 && repl.Slice.OpenEnd == 0 {
		slice := model.EmptySlice
		if s.Slice.Size()+repl.Slice.Size() != 0 {
			slice = model.NewSlice(s.Slice.Content.Append(repl.Slice.Content), s.Slice.OpenStart, repl.Slice.OpenEnd)
		}
		return NewReplaceStep(s.From, s.To+repl.To-repl.From, slice, s.Structure), true
	}
	if repl.To == s.From && repl.Slice.OpenStart == 0 && s.Slice.OpenEnd == 0 {
		slice := model.EmptySlice
		if s.Slice.Size()+repl.Slice.Size() != 0 {
			slice = model.NewSlice(repl.Slice.Content.Append(s.Slice.Content), repl.Slice.OpenStart, s.Slice.OpenEnd)
		}
		return NewReplaceStep(repl.From, s.To, slice, s.Structure), true
	}
	return nil, false
}

// ToJSON is a method of the Step interface.
func (s *ReplaceStep) ToJSON() map[string]interface{} {
	obj := map[string]interface{}{
		"stepType": "replace",
		"from":     s.From,
		"to":       s.To,
	}
	if s.Slice.Size() > 0 {
		obj["slice"] = s.Slice.ToJSON()
	}
	if s.Structure {
		obj["structure"] = true
	}
	return obj
}

// ReplaceStepFromJSON builds an RemoveMarkStep from a JSON representation.
func ReplaceStepFromJSON(schema *model.Schema, obj map[string]interface{}) (Step, error) {
	var from, to int
	switch f := obj["from"].(type) {
	case int:
		from = f
	case float64:
		from = int(f)
	default:
		return nil, errors.New("Invalid input for ReplaceStep.fromJSON")
	}
	switch t := obj["to"].(type) {
	case int:
		to = t
	case float64:
		to = int(t)
	default:
		return nil, errors.New("Invalid input for ReplaceStep.fromJSON")
	}
	raw, _ := obj["slice"].(map[string]interface{})
	slice, err := model.SliceFromJSON(schema, raw)
	if err != nil {
		return nil, err
	}
	structure, _ := obj["structure"].(bool)
	return NewReplaceStep(from, to, slice, structure), nil
}

// TableSortStepFromJSON builds a step that does nothing. It is used by atlaskit.
// Cf https://bitbucket.org/atlassian/atlaskit-mk-2/src/master/packages/editor/editor-core/src/plugins/table/utils/sort-step.ts
func TableSortStepFromJSON(schema *model.Schema, obj map[string]interface{}) (Step, error) {
	return NewReplaceStep(0, 0, model.EmptySlice), nil
}

var _ Step = &ReplaceStep{}

// ReplaceAroundStep replaces a part of the document with a slice of content,
// but preserve a range of the replaced content by moving it into the slice.
type ReplaceAroundStep struct {
	From      int
	To        int
	GapFrom   int
	GapTo     int
	Slice     *model.Slice
	Insert    int
	Structure bool
}

// NewReplaceAroundStep creates a replace-around step with the given range and
// gap. insert should be the point in the slice into which the content of the
// gap should be moved. structure has the same meaning as it has in the
// ReplaceStep class.
func NewReplaceAroundStep(from, to, gapFrom, gapTo int, slice *model.Slice, insert int, structure bool) *ReplaceAroundStep {
	return &ReplaceAroundStep{from, to, gapFrom, gapTo, slice, insert, structure}
}

// Apply is a method of the Step interface.
func (s *ReplaceAroundStep) Apply(doc *model.Node) StepResult {
	if s.Structure && (contentBetween(doc, s.From, s.GapFrom) || contentBetween(doc, s.GapTo, s.To)) {
		return Fail("Structure gap-replace would overwrite content")
	}

	gap, err := doc.Slice(s.GapFrom, s.GapTo)
	if err != nil {
		return Fail(err.Error())
	}
	if gap.OpenStart != 0 && gap.OpenEnd != 0 {
		return Fail("Gap is not a flat range")
	}
	inserted := s.Slice.InsertAt(s.Insert, gap.Content)
	if inserted == nil {
		return Fail("Content does not fit in gap")
	}
	return FromReplace(doc, s.From, s.To, inserted)
}

// GetMap is a method of the Step interface.
func (s *ReplaceAroundStep) GetMap() *StepMap {
	return NewStepMap([]int{s.From, s.GapFrom - s.From, s.Insert,
		s.GapTo, s.To - s.GapTo, s.Slice.Size() - s.Insert})
}

// Invert is a method of the Step interface.
func (s *ReplaceAroundStep) Invert(doc *model.Node) Step {
	gap := s.GapTo - s.GapFrom
	slice, err := doc.Slice(s.From, s.To)
	if err != nil {
		return nil
	}
	removed, err := slice.RemoveBetween(s.GapFrom-s.From, s.GapTo-s.To)
	if err != nil {
		return nil
	}
	return NewReplaceAroundStep(s.From, s.From+s.Slice.Size()+gap,
		s.From+s.Insert, s.From+s.Insert+gap,
		removed, s.GapFrom-s.From, s.Structure)
}

// Map is a method of the Step interface.
func (s *ReplaceAroundStep) Map(mapping Mappable) Step {
	from := mapping.MapResult(s.From, 1)
	to := mapping.MapResult(s.To, -1)
	gapFrom := mapping.Map(s.GapFrom, -1)
	gapTo := mapping.Map(s.GapTo, 1)
	if from.Deleted && to.Deleted || gapFrom < from.Pos || gapTo > to.Pos {
		return nil
	}
	return NewReplaceAroundStep(from.Pos, to.Pos, gapFrom, gapTo, s.Slice, s.Insert, s.Structure)
}

// Merge is a method of the Step interface.
func (s *ReplaceAroundStep) Merge(other Step) (Step, bool) {
	return nil, false
}

// ToJSON is a method of the Step interface.
func (s *ReplaceAroundStep) ToJSON() map[string]interface{} {
	obj := map[string]interface{}{
		"stepType": "replaceAround",
		"from":     s.From,
		"to":       s.To,
		"gapFrom":  s.GapFrom,
		"gapTo":    s.GapTo,
		"insert":   s.Insert,
	}
	if s.Slice.Size() > 0 {
		obj["slice"] = s.Slice.ToJSON()
	}
	if s.Structure {
		obj["structure"] = true
	}
	return obj
}

// ReplaceAroundStepFromJSON builds an RemoveMarkStep from a JSON representation.
func ReplaceAroundStepFromJSON(schema *model.Schema, obj map[string]interface{}) (Step, error) {
	var from, to, gapFrom, gapTo, insert int
	switch f := obj["from"].(type) {
	case int:
		from = f
	case float64:
		from = int(f)
	default:
		return nil, errors.New("Invalid input for ReplaceAroundStep.fromJSON")
	}
	switch t := obj["to"].(type) {
	case int:
		to = t
	case float64:
		to = int(t)
	default:
		return nil, errors.New("Invalid input for ReplaceAroundStep.fromJSON")
	}
	switch f := obj["gapFrom"].(type) {
	case int:
		gapFrom = f
	case float64:
		gapFrom = int(f)
	default:
		return nil, errors.New("Invalid input for ReplaceAroundStep.fromJSON")
	}
	switch t := obj["gapTo"].(type) {
	case int:
		gapTo = t
	case float64:
		gapTo = int(t)
	default:
		return nil, errors.New("Invalid input for ReplaceAroundStep.fromJSON")
	}
	switch n := obj["insert"].(type) {
	case int:
		insert = n
	case float64:
		insert = int(n)
	default:
		return nil, errors.New("Invalid input for ReplaceAroundStep.fromJSON")
	}
	raw, _ := obj["slice"].(map[string]interface{})
	slice, err := model.SliceFromJSON(schema, raw)
	if err != nil {
		return nil, err
	}
	structure, _ := obj["structure"].(bool)
	return NewReplaceAroundStep(from, to, gapFrom, gapTo, slice, insert, structure), nil
}

var _ Step = &ReplaceAroundStep{}

func contentBetween(doc *model.Node, from, to int) bool {
	dfrom, err := doc.Resolve(from)
	if err != nil {
		return true
	}
	dist := to - from
	depth := dfrom.Depth
	for dist > 0 && depth > 0 && dfrom.IndexAfter(depth) == dfrom.Node(depth).ChildCount() {
		depth--
		dist--
	}
	if dist > 0 {
		next := dfrom.Node(depth).MaybeChild(dfrom.IndexAfter(depth))
		for dist > 0 {
			if next == nil || next.IsLeaf() {
				return true
			}
			next = next.FirstChild()
			dist--
		}
	}
	return false
}
