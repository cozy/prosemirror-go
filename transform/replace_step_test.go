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

func TestReplaceBackspaceWithAccent(t *testing.T) {
	testDoc := doc(p("Numéro")).Node

	step1 := NewReplaceStep(6, 7, model.EmptySlice, false)
	step2 := NewReplaceStep(5, 6, model.EmptySlice, false)

	result := step1.Apply(testDoc)
	assert.Empty(t, result.Failed)
	assert.Equal(t, "Numér", *result.Doc.Content.Content[0].Content.Content[0].Text)
	result = step2.Apply(result.Doc)
	assert.Empty(t, result.Failed)
	assert.Equal(t, "Numé", *result.Doc.Content.Content[0].Content.Content[0].Text)
}
