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

func TestReplaceTwice(t *testing.T) {
	yes := func(from1, to1 int, txt1, expected1 string, from2, to2 int, txt2, expected2 string) {
		testDoc := doc(p("Numéro")).Node

		slice1 := model.EmptySlice
		if txt1 != "" {
			node := schema.Text(txt1, nil)
			fragment := model.NewFragment([]*model.Node{node})
			slice1 = &model.Slice{Content: fragment}
		}
		step1 := NewReplaceStep(from1, to1, slice1, false)
		result := step1.Apply(testDoc)
		assert.Empty(t, result.Failed)
		assert.Equal(t, expected1, *result.Doc.Content.Content[0].Content.Content[0].Text)

		slice2 := model.EmptySlice
		if txt2 != "" {
			node := schema.Text(txt2, nil)
			fragment := model.NewFragment([]*model.Node{node})
			slice2 = &model.Slice{Content: fragment}
		}
		step2 := NewReplaceStep(from2, to2, slice2, false)
		result = step2.Apply(result.Doc)
		assert.Empty(t, result.Failed)
		assert.Equal(t, expected2, *result.Doc.Content.Content[0].Content.Content[0].Text)
	}

	// Double backspace
	yes(6, 7, "", "Numér", 5, 6, "", "Numé")

	// An emoji in JS counts as 2 UTF-16 code units
	yes(2, 2, "👥", "N👥uméro", 4, 4, "🔎", "N👥🔎uméro")
}
