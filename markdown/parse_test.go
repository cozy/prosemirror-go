package markdown

import (
	"testing"

	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

var (
	// schema     = builder.Schema
	doc        = builder.Doc
	blockquote = builder.Blockquote
	h1         = builder.H1
	h2         = builder.H2
	p          = builder.P
	ol         = builder.Ol
	ul         = builder.Ul
	li         = builder.Li
	a          = builder.A
	pre        = builder.Pre
	// img        = builder.Img
	br     = builder.Br
	hr     = builder.Hr
	em     = builder.Em
	strong = builder.Strong
	code   = builder.Code
)

func TestMarkdown(t *testing.T) {
	parse := func(text string, doc builder.NodeWithTag) {
		// TODO
		// actual := DefaultParser.parse(text)
		// expected := doc.Node
		// assert.True(t, actual.Eq(expected), "%s != %s\n", actual.String(), expected.String())
	}

	serialize := func(doc builder.NodeWithTag, text string) {
		assert.Equal(t, DefaultSerializer.Serialize(doc.Node), text)
	}

	same := func(text string, doc builder.NodeWithTag) {
		parse(text, doc)
		serialize(doc, text)
	}

	// parses a paragraph
	same("hello!",
		doc(p("hello!")))

	// parses headings
	same("# one\n\n## two\n\nthree",
		doc(h1("one"), h2("two"), p("three")))

	// parses a blockquote
	same("> once\n\n> > twice",
		doc(blockquote(p("once")), blockquote(blockquote(p("twice")))))

	// FIXME bring back testing for preserving bullets and tight attrs
	// when supported again

	// parses a bullet list
	same("* foo\n\n  * bar\n\n  * baz\n\n* quux",
		doc(ul(li(p("foo"), ul(li(p("bar")), li(p("baz")))), li(p("quux")))))

	// parses an ordered list
	same("1. Hello\n\n2. Goodbye\n\n3. Nest\n\n   1. Hey\n\n   2. Aye",
		doc(ol(li(p("Hello")), li(p("Goodbye")), li(p("Nest"), ol(li(p("Hey")), li(p("Aye")))))))

	// parses a code block
	// TODO
	// same("Some code:\n\n```\nHere it is\n```\n\nPara",
	//      doc(p("Some code:"), schema.node("code_block", {params: ""}, [schema.text("Here it is")]), p("Para")))

	// parses an intended code block
	parse("Some code:\n\n    Here it is\n\nPara",
		doc(p("Some code:"), pre("Here it is"), p("Para")))

	// parses a fenced code block with info string
	// TODO
	// same("foo\n\n```javascript\n1\n```",
	//      doc(p("foo"), schema.node("code_block", {params: "javascript"}, [schema.text("1")])))

	// parses inline marks
	same("Hello. Some *em* text, some **strong** text, and some `code`",
		doc(p("Hello. Some ", em("em"), " text, some ", strong("strong"), " text, and some ", code("code"))))

	// parses overlapping inline marks
	// TODO
	// same("This is **strong *emphasized text with `code` in* it**",
	// 	doc(p("This is ", strong("strong ", em("emphasized text with ", code("code"), " in"), " it"))))

	// parses links inside strong text
	// TODO
	// same("**[link](foo) is bold**",
	// 	doc(p(strong(a("link"), " is bold"))))

	// parses code mark inside strong text
	same("**`code` is bold**",
		doc(p(strong(code("code"), " is bold"))))

	// parses code mark containing backticks
	same("``` one backtick: ` two backticks: `` ```",
		doc(p(code("one backtick: ` two backticks: ``"))))

	// parses code mark containing only whitespace
	serialize(doc(p("Three spaces: ", code("   "))),
		"Three spaces: `   `")

	// parses links
	same("My [link](foo) goes to foo",
		doc(p("My ", a("link"), " goes to foo")))

	// parses urls
	// TODO
	// same("Link to <https://prosemirror.net>",
	//      doc(p("Link to ", link({href: "https://prosemirror.net"}, "https://prosemirror.net"))))

	// parses emphasized urls
	// TODO
	// same("Link to *<https://prosemirror.net>*",
	//      doc(p("Link to ", em(link({href: "https://prosemirror.net"}, "https://prosemirror.net")))))

	// parses an image
	// TODO
	// same("Here's an image: ![x](img.png)",
	// 	doc(p("Here's an image: ", img)))

	// parses a line break
	same("line one\\\nline two",
		doc(p("line one", br, "line two")))

	// parses a horizontal rule
	same("one two\n\n---\n\nthree",
		doc(p("one two"), hr, p("three")))

	// ignores HTML tags
	// TODO
	// same("Foo < img> bar",
	// 	doc(p("Foo < img> bar")))

	// doesn't accidentally generate list markup
	same("1\\. foo",
		doc(p("1. foo")))

	// doesn't fail with line break inside inline mark
	same("**text1\ntext2**", doc(p(strong("text1\ntext2"))))

	// drops trailing hard breaks
	serialize(doc(p("a", br, br)), "a")

	// expels enclosing whitespace from inside emphasis
	serialize(doc(p("Some emphasized text with", strong(em("  whitespace   ")), "surrounding the emphasis.")),
		"Some emphasized text with  ***whitespace***   surrounding the emphasis.")

	// drops nodes when all whitespace is expelled from them
	// TODO
	// serialize(doc(p("Text with", em(" "), "an emphasized space")),
	// 	"Text with an emphasized space")

	// doesn't put a code block after a list item inside the list item
	same("* list item\n\n```\ncode\n```",
		doc(ul(li(p("list item"))), pre("code")))

	// doesn't escape characters in code
	same("foo`*`", doc(p("foo", code("*"))))
}
