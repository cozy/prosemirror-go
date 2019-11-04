package model_test

import (
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func TestNodeReplace(t *testing.T) {
	rpl := func(doc, insert, expect builder.NodeWithTag) {
		expected := expect.Node
		slice := EmptySlice
		if insert.Node != nil {
			slice = insert.Slice(insert.Tag["a"], insert.Tag["b"])
		}
		actual, err := doc.Replace(doc.Tag["a"], doc.Tag["b"], slice)
		if assert.NoError(t, err) {
			assert.True(t, actual.Eq(expected), "%s != %s\n", actual.String(), expected.String())
		}
	}

	// joins on delete
	rpl(doc(p("on<a>e"), p("t<b>wo")), builder.NodeWithTag{}, doc(p("onwo")))

	// merges matching blocks
	rpl(doc(p("on<a>e"), p("t<b>wo")), doc(p("xx<a>xx"), p("yy<b>yy")), doc(p("onxx"), p("yywo")))

	// merges when adding text
	rpl(doc(p("on<a>e"), p("t<b>wo")),
		doc(p("<a>H<b>")),
		doc(p("onHwo")))

	// can insert text
	rpl(doc(p("before"), p("on<a><b>e"), p("after")),
		doc(p("<a>H<b>")),
		doc(p("before"), p("onHe"), p("after")))

	// doesn't merge non-matching blocks
	rpl(doc(p("on<a>e"), p("t<b>wo")),
		doc(h1("<a>H<b>")),
		doc(p("onHwo")))

	// can merge a nested node
	rpl(doc(blockquote(blockquote(p("on<a>e"), p("t<b>wo")))),
		doc(p("<a>H<b>")),
		doc(blockquote(blockquote(p("onHwo")))))

	// can replace within a block
	rpl(doc(blockquote(p("a<a>bc<b>d"))),
		doc(p("x<a>y<b>z")),
		doc(blockquote(p("ayd"))))

	// can insert a lopsided slice
	rpl(doc(blockquote(blockquote(p("on<a>e"), p("two"), "<b>", p("three")))),
		doc(blockquote(p("aa<a>aa"), p("bb"), p("cc"), "<b>", p("dd"))),
		doc(blockquote(blockquote(p("onaa"), p("bb"), p("cc"), p("three")))))

	// can insert a deep, lopsided slice
	rpl(doc(blockquote(blockquote(p("on<a>e"), p("two"), p("three")), "<b>", p("x"))),
		doc(blockquote(p("aa<a>aa"), p("bb"), p("cc")), "<b>", p("dd")),
		doc(blockquote(blockquote(p("onaa"), p("bb"), p("cc")), p("x"))))

	// can merge multiple levels
	rpl(doc(blockquote(blockquote(p("hell<a>o"))), blockquote(blockquote(p("<b>a")))),
		builder.NodeWithTag{},
		doc(blockquote(blockquote(p("hella")))))

	// can merge multiple levels while inserting
	rpl(doc(blockquote(blockquote(p("hell<a>o"))), blockquote(blockquote(p("<b>a")))),
		doc(p("<a>i<b>")),
		doc(blockquote(blockquote(p("hellia")))))

	// can insert a split
	rpl(doc(p("foo<a><b>bar")),
		doc(p("<a>x"), p("y<b>")),
		doc(p("foox"), p("ybar")))

	// can insert a deep split
	rpl(doc(blockquote(p("foo<a>x<b>bar"))),
		doc(blockquote(p("<a>x")), blockquote(p("y<b>"))),
		doc(blockquote(p("foox")), blockquote(p("ybar"))))

	// can add a split one level up
	rpl(doc(blockquote(p("foo<a>u"), p("v<b>bar"))),
		doc(blockquote(p("<a>x")), blockquote(p("y<b>"))),
		doc(blockquote(p("foox")), blockquote(p("ybar"))))

	// keeps the node type of the left node
	rpl(doc(h1("foo<a>bar"), "<b>"),
		doc(p("foo<a>baz"), "<b>"),
		doc(h1("foobaz")))

	// keeps the node type even when empty
	rpl(doc(h1("<a>bar"), "<b>"),
		doc(p("foo<a>baz"), "<b>"),
		doc(h1("baz")))

	bad := func(doc, insert builder.NodeWithTag, pattern string) {
		slice := EmptySlice
		if insert.Node != nil {
			slice = insert.Slice(insert.Tag["a"], insert.Tag["b"])
		}
		_, err := doc.Replace(doc.Tag["a"], doc.Tag["b"], slice)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), pattern)
		}
	}

	// doesn't allow the left side to be too deep
	bad(doc(p("<a><b>")),
		doc(blockquote(p("<a>")), "<b>"),
		"deeper")

	// doesn't allow a depth mismatch
	bad(doc(p("<a><b>")),
		doc("<a>", p("<b>")),
		"Inconsistent")

	// rejects a bad fit
	bad(doc("<a><b>"),
		doc(p("<a>foo<b>")),
		"Invalid content")

	// rejects unjoinable content
	bad(doc(ul(li(p("a")), "<a>"), "<b>"),
		doc(p("foo", "<a>"), "<b>"),
		"Cannot join")

	// rejects an unjoinable delete
	bad(doc(blockquote(p("a"), "<a>"), ul("<b>", li(p("b")))),
		builder.NodeWithTag{},
		"Cannot join")

	// check content validity
	bad(doc(blockquote("<a>", p("hi")), "<b>"),
		doc(blockquote("hi", "<a>"), "<b>"),
		"Invalid content")
}
