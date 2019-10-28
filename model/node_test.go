package model_test

import (
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func TestNodeString(t *testing.T) {
	// nests
	assert.Equal(t,
		doc(ul(li(p("hey"), p()), li(p("foo")))).String(),
		`doc(bullet_list(list_item(paragraph("hey"), paragraph), list_item(paragraph("foo"))))`,
	)

	// shows inline children
	assert.Equal(t,
		doc(p("foo", img, br, "bar")).String(),
		`doc(paragraph("foo", image, hard_break, "bar"))`,
	)

	// shows marks
	assert.Equal(t,
		doc(p("foo", em("bar", strong("quux")), code("baz"))).String(),
		`doc(paragraph("foo", em("bar"), em(strong("quux")), code("baz")))`,
	)
}

func TestNodeCut(t *testing.T) {
	cut := func(doc, c builder.NodeWithTag) {
		expected := c.Node
		var actual *Node
		if b, ok := doc.Tag["b"]; ok {
			actual = doc.Cut(doc.Tag["a"], b)
		} else {
			actual = doc.Cut(doc.Tag["a"])
		}
		assert.True(t, actual.Eq(expected), "%s != %s\n", actual.String(), expected.String())
	}

	// extracts a full block
	cut(doc(p("foo"), "<a>", p("bar"), "<b>", p("baz")),
		doc(p("bar")))

	// cuts text
	cut(doc(p("0"), p("foo<a>bar<b>baz"), p("2")),
		doc(p("bar")))

	// cuts deeply
	cut(doc(blockquote(ul(li(p("a"), p("b<a>c")), li(p("d")), "<b>", li(p("e"))), p("3"))),
		doc(blockquote(ul(li(p("c")), li(p("d"))))))

	// works from the left
	cut(doc(blockquote(p("foo<b>bar"))),
		doc(blockquote(p("foo"))))

	// works to the right
	cut(doc(blockquote(p("foo<a>bar"))),
		doc(blockquote(p("bar"))))

	// preserves marks
	cut(doc(p("foo", em("ba<a>r", img, strong("baz"), br), "qu<b>ux", code("xyz"))),
		doc(p(em("r", img, strong("baz"), br), "qu")))
}

func TestNodesBetween(t *testing.T) {
	between := func(doc builder.NodeWithTag, nodes ...string) {
		i := 0
		doc.NodesBetween(doc.Tag["a"], doc.Tag["b"], func(node *Node, pos int, _ *Node, _ int) bool {
			if !assert.NotEqual(t, i, len(nodes), "More nodes iterated than listed ("+node.Type.Name+")") {
				compare := node.Type.Name
				if node.IsText() {
					compare = *node.Text
				}
				actual := nodes[i]
				i++
				assert.Equal(t, compare, actual)
				if !node.IsText() {
					assert.Equal(t, doc.NodeAt(pos), node)
				}
				return true
			}
			return false
		})
	}

	// iterates over text
	between(doc(p("foo<a>bar<b>baz")),
		"paragraph", "foobarbaz")

	// descends multiple levels
	between(doc(blockquote(ul(li(p("f<a>oo")), p("b"), "<b>"), p("c"))),
		"blockquote", "bullet_list", "list_item", "paragraph", "foo", "paragraph", "b")

	// iterates over inline nodes
	between(doc(p(em("x"), "f<a>oo", em("bar", img, strong("baz"), br), "quux", code("xy<b>z"))),
		"paragraph", "foo", "bar", "image", "baz", "hard_break", "quux", "xyz")
}

func TestNodeTextContent(t *testing.T) {
	// works on a whole doc
	assert.Equal(t, doc(p("foo")).TextContent(), "foo")

	// works on a text node
	assert.Equal(t, schema.Text("foo").TextContent(), "foo")

	// works on a nested element
	assert.Equal(t,
		doc(ul(li(p("hi")), li(p(em("a"), "b")))).TextContent(),
		"hiab")
}
