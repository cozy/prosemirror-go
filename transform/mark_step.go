package transform

import (
	"errors"

	"github.com/cozy/prosemirror-go/model"
)

type mapFn func(node, parent *model.Node) *model.Node

func mapFragment(fragment *model.Fragment, f mapFn, parent *model.Node) (*model.Fragment, error) {
	var mapped []*model.Node
	for i := 0; i < fragment.ChildCount(); i++ {
		child, err := fragment.Child(i)
		if err != nil {
			return nil, err
		}
		if child.Content.Size > 0 {
			copied, err := mapFragment(child.Content, f, child)
			if err != nil {
				return nil, err
			}
			child = child.Copy(copied)
		}
		if child.IsInline() {
			child = f(child, parent)
		}
		mapped = append(mapped, child)
	}
	content, err := model.FragmentFrom(mapped)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// AddMarkStep adds a mark to all inline content between two positions.
type AddMarkStep struct {
	From int
	To   int
	Mark *model.Mark
}

// NewAddMarkStep is the constructor for AddMarkStep.
func NewAddMarkStep(from, to int, mark *model.Mark) *AddMarkStep {
	return &AddMarkStep{From: from, To: to, Mark: mark}
}

// Apply is a method of the Step interface.
func (s *AddMarkStep) Apply(doc *model.Node) StepResult {
	oldSlice, err := doc.Slice(s.From, s.To)
	if err != nil {
		return Fail(err.Error())
	}
	dFrom, err := doc.Resolve(s.From)
	if err != nil {
		return Fail(err.Error())
	}
	parent := dFrom.Node(dFrom.SharedDepth(s.To))
	fragment, err := mapFragment(oldSlice.Content, func(node, parent *model.Node) *model.Node {
		if !parent.Type.AllowsMarkType(s.Mark.Type) {
			return node
		}
		return node.Mark(s.Mark.AddToSet(node.Marks))
	}, parent)
	if err != nil {
		return Fail(err.Error())
	}
	slice := model.NewSlice(fragment, oldSlice.OpenStart, oldSlice.OpenEnd)
	return FromReplace(doc, s.From, s.To, slice)
}

// GetMap is a method of the Step interface.
func (s *AddMarkStep) GetMap() *StepMap {
	return EmptyStepMap
}

// Invert is a method of the Step interface.
func (s *AddMarkStep) Invert(doc *model.Node) Step {
	return NewRemoveMarkStep(s.From, s.To, s.Mark)
}

// Map is a method of the Step interface.
func (s *AddMarkStep) Map(mapping Mappable) Step {
	from := mapping.MapResult(s.From, 1)
	to := mapping.MapResult(s.To, -1)
	if from.Deleted && to.Deleted || from.Pos >= to.Pos {
		return nil
	}
	return NewAddMarkStep(from.Pos, to.Pos, s.Mark)
}

// Merge is a method of the Step interface.
func (s *AddMarkStep) Merge(other Step) (Step, bool) {
	add, ok := other.(*AddMarkStep)
	if ok && add.Mark.Eq(s.Mark) && s.From <= add.To && s.To >= add.From {
		f := s.From
		if f > add.From {
			f = add.From
		}
		t := s.To
		if t < add.To {
			t = add.To
		}
		return NewAddMarkStep(f, t, s.Mark), true
	}
	return nil, false
}

// ToJSON is a method of the Step interface.
func (s *AddMarkStep) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"stepType": "addMark",
		"mark":     s.Mark.ToJSON(),
		"from":     s.From,
		"to":       s.To,
	}
}

// AddMarkStepFromJSON builds an AddMarkStep from a JSON representation.
func AddMarkStepFromJSON(schema *model.Schema, obj map[string]interface{}) (Step, error) {
	raw, ok := obj["mark"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Invalid input for AddMarkStep.fromJSON")
	}
	mark, err := model.MarkFromJSON(schema, raw)
	if err != nil {
		return nil, err
	}
	var from, to int
	switch f := obj["from"].(type) {
	case int:
		from = f
	case float64:
		from = int(f)
	}
	switch t := obj["to"].(type) {
	case int:
		to = t
	case float64:
		to = int(t)
	}
	return NewAddMarkStep(from, to, mark), nil
}

var _ Step = &AddMarkStep{}

// RemoveMarkStep adds a mark to all inline content between two positions.
type RemoveMarkStep struct {
	From int
	To   int
	Mark *model.Mark
}

// NewRemoveMarkStep is the constructor for RemoveMarkStep.
func NewRemoveMarkStep(from, to int, mark *model.Mark) *RemoveMarkStep {
	return &RemoveMarkStep{From: from, To: to, Mark: mark}
}

// Apply is a method of the Step interface.
func (s *RemoveMarkStep) Apply(doc *model.Node) StepResult {
	oldSlice, err := doc.Slice(s.From, s.To)
	if err != nil {
		return Fail(err.Error())
	}
	fragment, err := mapFragment(oldSlice.Content, func(node, parent *model.Node) *model.Node {
		return node.Mark(s.Mark.RemoveFromSet(node.Marks))
	}, nil)
	if err != nil {
		return Fail(err.Error())
	}
	slice := model.NewSlice(fragment, oldSlice.OpenStart, oldSlice.OpenEnd)
	return FromReplace(doc, s.From, s.To, slice)
}

// GetMap is a method of the Step interface.
func (s *RemoveMarkStep) GetMap() *StepMap {
	return EmptyStepMap
}

// Invert is a method of the Step interface.
func (s *RemoveMarkStep) Invert(doc *model.Node) Step {
	return NewAddMarkStep(s.From, s.To, s.Mark)
}

// Map is a method of the Step interface.
func (s *RemoveMarkStep) Map(mapping Mappable) Step {
	from := mapping.MapResult(s.From, 1)
	to := mapping.MapResult(s.To, -1)
	if from.Deleted && to.Deleted || from.Pos >= to.Pos {
		return nil
	}
	return NewRemoveMarkStep(from.Pos, to.Pos, s.Mark)
}

// Merge is a method of the Step interface.
func (s *RemoveMarkStep) Merge(other Step) (Step, bool) {
	rem, ok := other.(*RemoveMarkStep)
	if ok && rem.Mark.Eq(s.Mark) && s.From <= rem.To && s.To >= rem.From {
		f := s.From
		if f > rem.From {
			f = rem.From
		}
		t := s.To
		if t < rem.To {
			t = rem.To
		}
		return NewRemoveMarkStep(f, t, s.Mark), true
	}
	return nil, false
}

// ToJSON is a method of the Step interface.
func (s *RemoveMarkStep) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"stepType": "removeMark",
		"mark":     s.Mark.ToJSON(),
		"from":     s.From,
		"to":       s.To,
	}
}

// RemoveMarkStepFromJSON builds an RemoveMarkStep from a JSON representation.
func RemoveMarkStepFromJSON(schema *model.Schema, obj map[string]interface{}) (Step, error) {
	raw, ok := obj["mark"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Invalid input for RemoveMarkStep.fromJSON")
	}
	mark, err := model.MarkFromJSON(schema, raw)
	if err != nil {
		return nil, err
	}
	var from, to int
	switch f := obj["from"].(type) {
	case int:
		from = f
	case float64:
		from = int(f)
	}
	switch t := obj["to"].(type) {
	case int:
		to = t
	case float64:
		to = int(t)
	}
	return NewRemoveMarkStep(from, to, mark), nil
}

var _ Step = &RemoveMarkStep{}
