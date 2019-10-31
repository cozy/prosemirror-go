package transform

import "github.com/cozy/prosemirror-go/model"

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
	oldSlice := doc.Slice(s.From, s.To)
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
	oldSlice := doc.Slice(s.From, s.To)
	fragment, err := mapFragment(oldSlice.Content, func(parent, node *model.Node) *model.Node {
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

var _ Step = &RemoveMarkStep{}
