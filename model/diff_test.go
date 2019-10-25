package model_test

import (
	"testing"

	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func TestFragmentFindDiffStart(t *testing.T) {
	start := func(a, b builder.NodeWithTag) {
		found := a.Content.FindDiffStart(b.Content)
		expected, ok := a.Tag["a"]
		if ok {
			if assert.NotNil(t, found) {
				assert.Equal(t, expected, *found)
			}
		} else {
			assert.Nil(t, found)
		}
	}

	// returns null for identical nodes
	start(
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
	)

	// notices when one node is longer
	start(
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye")), "<a>"),
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye")), p("oops")),
	)

	// notices when one node is shorter
	start(
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye")), "<a>", p("oops")),
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
	)

	// notices differing marks
	start(
		doc(p("a<a>", em("b"))),
		doc(p("a", strong("b"))),
	)

	// stops at longer text
	start(
		doc(p("foo<a>bar", em("b"))),
		doc(p("foo", em("b"))),
	)

	// stops at a different character
	start(
		doc(p("foo<a>bar")),
		doc(p("foocar")),
	)

	// stops at a different node type
	start(
		doc(p("a"), "<a>", p("b")),
		doc(p("a"), h1("b")),
	)

	// works when the difference is at the start
	start(
		doc("<a>", p("b")),
		doc(h1("b")),
	)

	// notices a different attribute
	start(
		doc(p("a"), "<a>", h1("foo")),
		doc(p("a"), h2("foo")),
	)
}

func TestFragmentFindDiffEnd(t *testing.T) {
	end := func(a, b builder.NodeWithTag) {
		found := a.Content.FindDiffEnd(b.Content)
		expected, ok := a.Tag["a"]
		if ok {
			if assert.NotNil(t, found) {
				assert.Equal(t, expected, found.A)
			}
		} else {
			assert.Nil(t, found)
		}
	}

	// returns null when there is no difference
	end(
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
	)

	// notices when the second doc is longer
	end(
		doc("<a>", p("a", em("b")), p("hello"), blockquote(h1("bye"))),
		doc(p("oops"), p("a", em("b")), p("hello"), blockquote(h1("bye"))),
	)

	// notices when the second doc is shorter
	end(doc(p("oops"), "<a>", p("a", em("b")), p("hello"), blockquote(h1("bye"))),
		doc(p("a", em("b")), p("hello"), blockquote(h1("bye"))),
	)

	// notices different styles
	end(doc(p("a", em("b"), "<a>c")),
		doc(p("a", strong("b"), "c")),
	)

	// spots longer text
	end(doc(p("bar<a>foo", em("b"))),
		doc(p("foo", em("b"))),
	)

	// spots different text
	end(doc(p("foob<a>ar")),
		doc(p("foocar")),
	)

	// notices different nodes
	end(doc(p("a"), "<a>", p("b")),
		doc(h1("a"), p("b")),
	)

	// notices a difference at the end
	end(doc(p("b"), "<a>"),
		doc(h1("b")),
	)

	// handles a similar start
	end(doc("<a>", p("hello")),
		doc(p("hey"), p("hello")),
	)
}
