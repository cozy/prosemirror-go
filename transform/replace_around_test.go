package transform

import (
	"testing"

	"github.com/cozy/prosemirror-go/model"
	"github.com/stretchr/testify/assert"
)

func TestReplaceAround(t *testing.T) {
	testDoc := doc(p("Ma super note")).Node

	frag := model.NewFragment([]*model.Node{h1().Node})
	slice := model.NewSlice(frag, 0, 0)
	step := NewReplaceAroundStep(0, 15, 1, 14, slice, 1, true)

	result := step.Apply(testDoc)
	assert.Empty(t, result.Failed)
}
