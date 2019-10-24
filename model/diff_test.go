package model_test

import (
	"testing"

	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func TestFragmentFindDiffStart(t *testing.T) {
	start := func(a, b builder.NodeWithTag) {
		found := a.Content.FindDiffStart(b.Content)
		if assert.NotNil(t, found) {
			assert.Equal(t, *found, a.Tag["a"])
		}
	}

	// returns null for identical nodes
	start(
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
	)
}
