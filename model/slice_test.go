package model_test

import (
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func TestNodeSlice(t *testing.T) {
	test := func(doc, expect builder.NodeWithTag, openStart, openEnd int) {
		var slice *Slice
		if b, ok := doc.Tag["b"]; ok {
			slice = doc.Slice(doc.Tag["a"], b)
		} else {
			slice = doc.Slice(doc.Tag["a"])
		}
		assert.True(t, slice.Content.Eq(expect.Content))
		assert.Equal(t, slice.OpenStart, openStart)
		assert.Equal(t, slice.OpenEnd, openEnd)
	}

	// can cut half a paragraph
	test(doc(p("hello<b> world")), doc(p("hello")), 0, 1)

	// can cut to the end of a pragraph
	test(doc(p("hello<b>")), doc(p("hello")), 0, 1)

	// leaves off extra content
	test(doc(p("hello<b> world"), p("rest")), doc(p("hello")), 0, 1)

	// preserves styles
	test(doc(p("hello ", em("WOR<b>LD"))), doc(p("hello ", em("WOR"))), 0, 1)

	// can cut multiple blocks
	test(doc(p("a"), p("b<b>")), doc(p("a"), p("b")), 0, 1)

	// can cut to a top-level position
	test(doc(p("a"), "<b>", p("b")), doc(p("a")), 0, 0)

	// can cut to a deep position
	test(doc(blockquote(ul(li(p("a")), li(p("b<b>"))))),
		doc(blockquote(ul(li(p("a")), li(p("b"))))), 0, 4)

	// can cut everything after a position
	test(doc(p("hello<a> world")), doc(p(" world")), 1, 0)

	// can cut from the start of a textblock
	test(doc(p("<a>hello")), doc(p("hello")), 1, 0)

	// leaves off extra content before
	test(doc(p("foo"), p("bar<a>baz")), doc(p("baz")), 1, 0)

	// preserves styles after cut
	test(doc(p("a sentence with an ", em("emphasized ", a("li<a>nk")), " in it")),
		doc(p(em(a("nk")), " in it")), 1, 0)

	// preserves styles started after cut
	test(doc(p("a ", em("sentence"), " wi<a>th ", em("text"), " in it")),
		doc(p("th ", em("text"), " in it")), 1, 0)

	// can cut from a top-level position
	test(doc(p("a"), "<a>", p("b")), doc(p("b")), 0, 0)

	// can cut from a deep position
	test(doc(blockquote(ul(li(p("a")), li(p("<a>b"))))),
		doc(blockquote(ul(li(p("b"))))), 4, 0)

	// can cut part of a text node
	test(doc(p("hell<a>o wo<b>rld")), p("o wo"), 0, 0)

	// can cut across paragraphs
	test(doc(p("on<a>e"), p("t<b>wo")), doc(p("e"), p("t")), 1, 1)

	// can cut part of marked text
	test(doc(p("here's noth<a>ing and ", em("here's e<b>m"))),
		p("ing and ", em("here's e")), 0, 0)

	// can cut across different depths
	test(doc(ul(li(p("hello")), li(p("wo<a>rld")), li(p("x"))), p(em("bo<b>o"))),
		doc(ul(li(p("rld")), li(p("x"))), p(em("bo"))), 3, 1)

	// can cut between deeply nested nodes
	test(doc(blockquote(p("foo<a>bar"), ul(li(p("a")), li(p("b"), "<b>", p("c"))), p("d"))),
		blockquote(p("bar"), ul(li(p("a")), li(p("b")))), 1, 2)

	// can include parents
	// TODO
	// let d = doc(blockquote(p("fo<a>o"), p("bar<b>")))
	// let slice = d.slice(d.tag.a, d.tag.b, true)
	// ist(slice.toString(), '<blockquote(paragraph("o"), paragraph("bar"))>(2,2)')
	// })
}
