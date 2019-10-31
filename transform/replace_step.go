package transform

import (
	"fmt"

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
	return NewReplaceStep(s.From, s.From+s.Slice.Size(), doc.Slice(s.From, s.To))
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
	if !ok || repl.Structure != s.Structure {
		fmt.Printf("ok = %v\n", ok)
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

var _ Step = &ReplaceStep{}

func contentBetween(doc *model.Node, from, to int) bool {
	dfrom, err := doc.Resolve(from)
	if err != nil {
		panic(err)
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
