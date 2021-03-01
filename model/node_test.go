package model_test

import (
	"testing"

	. "github.com/shodgson/prosemirror-go/model"
	"github.com/shodgson/prosemirror-go/test/builder"
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

func TestNodeFrom(t *testing.T) {
	from := func(arg interface{}, expect builder.NodeWithTag) {
		expected := expect.Node
		fragment, err := FragmentFrom(arg)
		assert.NoError(t, err)
		actual := expect.Copy(fragment)
		assert.True(t, actual.Eq(expected), "%s != %s\n", actual.String(), expected.String())
	}

	// wraps a single node
	para, err := schema.Node("paragraph")
	assert.NoError(t, err)
	from(para, doc(p()))

	// wraps an array
	hard, err := schema.Node("hard_break")
	assert.NoError(t, err)
	from([]*Node{hard, schema.Text("foo")}, p(br, "foo"))

	// preserves a fragment
	from(doc(p("foo")).Content, doc(p("foo")))

	// accepts null
	from(nil, p())

	// joins adjacent text
	from([]*Node{schema.Text("a"), schema.Text("b")}, p("ab"))
}

func TestNodeToJSON(t *testing.T) {
	roundTrip := func(doc builder.NodeWithTag) {
		result, err := NodeFromJSON(schema, doc.ToJSON())
		assert.NoError(t, err)
		assert.True(t, result.Eq(doc.Node))
	}

	// can serialize a simple node
	roundTrip(doc(p("foo")))

	// can serialize marks
	roundTrip(doc(p("foo", em("bar", strong("baz")), " ", a("x"))))

	// can serialize inline leaf nodes
	roundTrip(doc(p("foo", em(img, "bar"))))

	// can serialize block leaf nodes
	roundTrip(doc(p("a"), hr, p("b"), p()))

	// can serialize nested nodes
	roundTrip(doc(blockquote(ul(li(p("a"), p("b")), li(p(img))), p("c")), p("d")))
}

func TestNodeToString(t *testing.T) {
	customSchema, err := NewSchema(&SchemaSpec{
		Nodes: []*NodeSpec{
			{Key: "doc", Content: "paragraph+"},
			{Key: "paragraph", Content: "text*"},
			{Key: "text", ToDebugString: func(_ *Node) string { return "custom_text" }},
			{Key: "hard_break", ToDebugString: func(_ *Node) string { return "custom_hard_break" }},
		},
	})
	assert.NoError(t, err)

	// should have the default toString method [text]
	assert.Equal(t, schema.Text("hello").String(), `"hello"`)

	// should have the default toString method [br]
	assert.Equal(t, br().String(), "hard_break")

	// should be able to redefine it from NodeSpec by specifying toDebugString method
	assert.Equal(t, customSchema.Text("hello").String(), "custom_text")

	// should be respected by Fragment
	hardBreak, err := customSchema.NodeType("hard_break")
	assert.NoError(t, err)
	hard, err := hardBreak.CreateChecked()
	assert.NoError(t, err)
	assert.Equal(t,
		FragmentFromArray(
			[]*Node{customSchema.Text("hello"), hard, customSchema.Text("world")},
		).String(),
		"<custom_text, custom_hard_break, custom_text>",
	)
}

func TestNodeReplaceText(t *testing.T) {
	replace := func(from, to int, text string, initial, expected builder.NodeWithTag) {
		txt := schema.Text(text, nil)
		fragment := NewFragment([]*Node{txt})
		slice := &Slice{Content: fragment}

		result, err := initial.Replace(from, to, slice)
		assert.NoError(t, err)
		assert.True(t, result.Eq(expected.Node))
	}

	replace(2, 2, "ô", doc(p("ab")), doc(p("aôb")))
	replace(1, 1, "ô", doc(p("ôô")), doc(p("ôôô")))
	replace(2, 2, "ô", doc(p("ôô")), doc(p("ôôô")))
	replace(3, 3, "ô", doc(p("ôô")), doc(p("ôôô")))
}

func TestNodeSize(t *testing.T) {
	nodeSize := func(node *Node, expected int) {
		assert.Equal(t, node.NodeSize(), expected)
	}

	nodeSize(schema.Text("a"), 1)
	nodeSize(schema.Text("hello world"), 11)
	nodeSize(schema.Text("ô"), 1)
	nodeSize(schema.Text("👥"), 2)
}

func TestNodeTextBetween(t *testing.T) {
	txt := schema.Text("hâhîhô", nil)
	assert.Equal(t, "hî", txt.TextBetween(2, 4))
}
