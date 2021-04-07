package transform

import (
	"errors"

	"github.com/cozy/prosemirror-go/model"
)

// SetAttrsStep can be used to change the attributes of a node.
//
// For more context, see:
// - https://discuss.prosemirror.net/t/preventing-image-placeholder-replacement-from-being-undone/1394/1
// - https://bitbucket.org/atlassian/atlassian-frontend-mirror/src/master/editor/adf-schema/src/steps/set-attrs.tsx
type SetAttrsStep struct {
	Pos   int
	Attrs map[string]interface{}
}

// NewSetAttrsStep is a constructor for SetAttrsStep
func NewSetAttrsStep(pos int, attrs map[string]interface{}) *SetAttrsStep {
	return &SetAttrsStep{Pos: pos, Attrs: attrs}
}

// Apply is a method of the Step interface.
func (s *SetAttrsStep) Apply(doc *model.Node) StepResult {
	target := doc.NodeAt(s.Pos)
	if target == nil {
		return Fail("No node at given position")
	}

	attrs := map[string]interface{}{}
	for k, v := range target.Attrs {
		attrs[k] = v
	}
	for k, v := range s.Attrs {
		attrs[k] = v
	}

	newNode, err := target.Type.Create(attrs, model.EmptyFragment, target.Marks)
	if err != nil {
		return Fail(err.Error())
	}
	leaf := 0
	if !target.IsLeaf() {
		leaf = 1
	}
	fragment, err := model.FragmentFrom(newNode)
	if err != nil {
		return Fail(err.Error())
	}
	slice := model.NewSlice(fragment, 0, leaf)
	return FromReplace(doc, s.Pos, s.Pos+1, slice)
}

// GetMap is a method of the Step interface.
func (s *SetAttrsStep) GetMap() *StepMap {
	return EmptyStepMap
}

// Invert is a method of the Step interface.
func (s *SetAttrsStep) Invert(doc *model.Node) Step {
	attrs := map[string]interface{}{}
	target := doc.NodeAt(s.Pos)
	if target != nil {
		attrs = target.Attrs
	}
	return NewSetAttrsStep(s.Pos, attrs)
}

// Map is a method of the Step interface.
func (s *SetAttrsStep) Map(mapping Mappable) Step {
	result := mapping.MapResult(s.Pos, 1)
	if result.Deleted {
		return nil
	}
	return NewSetAttrsStep(result.Pos, s.Attrs)
}

// Merge is a method of the Step interface.
func (s *SetAttrsStep) Merge(other Step) (Step, bool) {
	return nil, false
}

// ToJSON is a method of the Step interface.
func (s *SetAttrsStep) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"stepType": "setAttrs",
		"pos":      s.Pos,
		"attrs":    s.Attrs,
	}
}

// SetAttrsStepFromJSON builds an SetAttrsStep from a JSON representation.
func SetAttrsStepFromJSON(schema *model.Schema, obj map[string]interface{}) (Step, error) {
	attrs, ok := obj["attrs"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Invalid input for SetAttrsStep.fromJSON")
	}
	var pos int
	switch p := obj["pos"].(type) {
	case int:
		pos = p
	case float64:
		pos = int(p)
	}
	return NewSetAttrsStep(pos, attrs), nil
}

var _ Step = &SetAttrsStep{}
