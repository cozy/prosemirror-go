package markdown

import (
	"testing"

	"github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/schema/basic"
	"github.com/cozy/prosemirror-go/schema/list"
	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

var (
	empty        = ""
	headingAttrs = map[string]*model.AttributeSpec{
		"level": {Default: 1.0},
	}
	codeAttrs = map[string]*model.AttributeSpec{
		"params": {Default: ""},
	}
	imageAttrs = map[string]*model.AttributeSpec{
		"src":   {},
		"alt":   {Default: nil},
		"title": {Default: nil},
	}
	nodes = []*model.NodeSpec{
		{Key: "doc", Content: "block+"},
		{Key: "paragraph", Content: "inline*", Group: "block"},
		{Key: "blockquote", Content: "block+", Group: "block"},
		{Key: "horizontal_rule", Group: "block"},
		{Key: "heading", Content: "inline*", Group: "block", Attrs: headingAttrs},
		{Key: "code_block", Content: "text*", Marks: &empty, Group: "block", Attrs: codeAttrs},
		{Key: "text", Group: "inline"},
		{Key: "image", Group: "inline", Inline: true, Attrs: imageAttrs},
		{Key: "hard_break", Group: "inline", Inline: true},
	}

	schema, _ = model.NewSchema(&model.SchemaSpec{
		Nodes: list.AddListNodes(nodes, "paragraph block*", "block"),
		Marks: basic.Schema.Spec.Marks,
	})
	out = builder.Builders(schema, map[string]builder.Spec{
		"p":   {"nodeType": "paragraph"},
		"h1":  {"nodeType": "heading", "level": 1},
		"h2":  {"nodeType": "heading", "level": 2},
		"hr":  {"nodeType": "horizontal_rule"},
		"li":  {"nodeType": "list_item"},
		"ol":  {"nodeType": "ordered_list"},
		"ol3": {"nodeType": "ordered_list", "order": float64(3)},
		"ul":  {"nodeType": "bullet_list"},
		"pre": {"nodeType": "code_block"},
		"a":   {"markType": "link", "href": "foo"},
		"br":  {"nodeType": "hard_break"},
		"img": {"nodeType": "image", "src": "img.png", "alt": "x"},
	})

	doc        = out["doc"].(builder.NodeBuilder)
	blockquote = out["blockquote"].(builder.NodeBuilder)
	p          = out["p"].(builder.NodeBuilder)
	h1         = out["h1"].(builder.NodeBuilder)
	h2         = out["h2"].(builder.NodeBuilder)
	// hr         = out["hr"].(builder.NodeBuilder)
	li     = out["li"].(builder.NodeBuilder)
	ol     = out["ol"].(builder.NodeBuilder)
	ol3    = out["ol3"].(builder.NodeBuilder)
	ul     = out["ul"].(builder.NodeBuilder)
	pre    = out["pre"].(builder.NodeBuilder)
	a      = out["a"].(builder.MarkBuilder)
	br     = out["br"].(builder.NodeBuilder)
	em     = out["em"].(builder.MarkBuilder)
	strong = out["strong"].(builder.MarkBuilder)
	code   = out["code"].(builder.MarkBuilder)
	img    = out["img"].(builder.NodeBuilder)
	link   = out["link"].(builder.MarkBuilder)
)

func TestMarkdown(t *testing.T) {
	parse := func(text string, doc builder.NodeWithTag) {
		parser := goldmark.DefaultParser()
		actual, err := ParseMarkdown(parser, DefaultNodeMapper, []byte(text), schema)
		require.NoError(t, err)
		expected := doc.Node
		require.True(t, actual.Eq(expected), "%s != %s\n", actual.String(), expected.String())
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

	// preserves ordered list start number
	same("3. Foo\n\n4. Bar",
		doc(ol3(li(p("Foo")), li(p("Bar")))))

	// parses a code block
	node, err := schema.Node("code_block", map[string]interface{}{"params": ""}, []interface{}{schema.Text("Here it is")})
	assert.NoError(t, err)
	same("Some code:\n\n```\nHere it is\n```\n\nPara",
		doc(p("Some code:"), node, p("Para")))

	// TODO parses an intended code block
	// parse("Some code:\n\n    Here it is\n\nPara",
	// 	doc(p("Some code:"), pre("Here it is"), p("Para")))

	// TODO parses a fenced code block with info string
	// node, err = schema.Node("code_block", map[string]interface{}{"params": "javascript"}, []interface{}{schema.Text("1")})
	// assert.NoError(t, err)
	// same("foo\n\n```javascript\n1\n```",
	// 	doc(p("foo"), node))

	// parses inline marks
	same("Hello. Some *em* text, some **strong** text, and some `code`",
		doc(p("Hello. Some ", em("em"), " text, some ", strong("strong"), " text, and some ", code("code"))))

	// parses overlapping inline marks
	same("This is **strong *emphasized text with `code` in* it**",
		doc(p("This is ", strong("strong ", em("emphasized text with ", code("code"), " in"), " it"))))

	// parses links inside strong text
	same("**[link](foo) is bold**",
		doc(p(strong(a("link"), " is bold"))))

	// parses emphasis inside links
	same("[link *foo **bar** `#`*](foo)",
		doc(p(a("link ", em("foo ", strong("bar"), " ", code("#"))))))

	// parses code mark inside strong text
	same("**`code` is bold**",
		doc(p(strong(code("code"), " is bold"))))

	// parses code mark containing backticks
	same("``` one backtick: ` two backticks: `` ```",
		doc(p(code("one backtick: ` two backticks: ``"))))

	// parses code mark containing only whitespace
	serialize(doc(p("Three spaces: ", code("   "))),
		"Three spaces: `   `")

	// parses hard breaks
	same("foo\\\nbar", doc(p("foo", br(), "bar")))
	same("*foo\\\nbar*", doc(p(em("foo", br(), "bar"))))

	// parses links
	same("My [link](foo) goes to foo",
		doc(p("My ", a("link"), " goes to foo")))

	// parses urls
	same("Link to <https://prosemirror.net>",
		doc(p("Link to ", link(map[string]interface{}{"href": "https://prosemirror.net"}, "https://prosemirror.net"))))

	// correctly serializes relative urls
	same("[foo.html](foo.html)",
		doc(p(link(map[string]interface{}{"href": "foo.html"}, "foo.html"))))

	// can handle link titles
	same(`[a](x.html "title \"quoted\"")`,
		doc(p(link(map[string]interface{}{"href": "x.html", "title": `title "quoted"`}, "a"))))

	// doesn't escape underscores in link
	same("[link](http://foo.com/a_b_c)",
		doc(p(link(map[string]interface{}{"href": "http://foo.com/a_b_c"}, "link"))))

	// parses emphasized urls
	same("Link to *<https://prosemirror.net>*",
		doc(p("Link to ", em(link(map[string]interface{}{"href": "https://prosemirror.net"}, "https://prosemirror.net")))))

	// TODO parses an image
	// same("Here's an image: ![x](img.png)",
	// 	doc(p("Here's an image: ", img)))

	// parses a line break
	same("line one\\\nline two",
		doc(p("line one", br, "line two")))

	// TODO parses a horizontal rule
	// same("one two\n\n---\n\nthree",
	// 	doc(p("one two"), hr, p("three")))

	// ignores HTML tags
	same("Foo < img> bar",
		doc(p("Foo < img> bar")))

	// escapes special characters
	same("Foo \\*bar",
		doc(p("Foo *bar")))

	// doesn't accidentally generate list markup
	same("1\\. foo",
		doc(p("1. foo")))

	// doesn't fail with line break inside inline mark
	serialize(doc(p(strong("text1\ntext2"))), "**text1\ntext2**")

	// drops trailing hard breaks
	serialize(doc(p("a", br, br)), "a")

	// expels enclosing whitespace from inside emphasis
	serialize(doc(p("Some emphasized text with", strong(em("  whitespace   ")), "surrounding the emphasis.")),
		"Some emphasized text with  ***whitespace***   surrounding the emphasis.")

	// drops nodes when all whitespace is expelled from them
	serialize(doc(p("Text with", em(" "), "an emphasized space")),
		"Text with an emphasized space")

	// doesn't put a code block after a list item inside the list item
	same("* list item\n\n```\ncode\n```",
		doc(ul(li(p("list item"))), pre("code")))

	// doesn't escape characters in code
	same("foo`*`", doc(p("foo", code("*"))))

	// doesn't escape underscores between word characters
	same("abc_def", doc(p("abc_def")))

	// doesn't escape strips of underscores between word characters
	same("abc___def", doc(p("abc___def")))

	// escapes underscores at word boundaries
	same("\\_abc\\_", doc(p("_abc_")))

	// escapes underscores surrounded by non-word characters
	same("/\\_abc\\_)", doc(p("/_abc_)")))

	// ensure no escapes in url
	parse("[text](https://example.com/_file/#~anchor)",
		doc(p(a(map[string]interface{}{"href": "https://example.com/_file/#~anchor"}, "text"))))

	// ensure no escapes in autolinks
	same("<https://example.com/_file/#~anchor>",
		doc(p(a(map[string]interface{}{"href": "https://example.com/_file/#~anchor"}, "https://example.com/_file/#~anchor"))))

	// escape ! in front of links
	serialize(doc(p("!", a("text"))),
		"\\![text](foo)")

	// escape of URL in links and images
	serialize(doc(p(a(map[string]interface{}{"href": "foo):"}, "link"))), "[link](foo\\):)")
	serialize(doc(p(a(map[string]interface{}{"href": "(foo"}, "link"))), "[link](\\(foo)")
	serialize(doc(p(img(map[string]interface{}{"src": "foo):"}))), "![x](foo\\):)")
	serialize(doc(p(img(map[string]interface{}{"src": "(foo"}))), "![x](\\(foo)")
	serialize(doc(p(a(map[string]interface{}{"title": "bar", "href": "foo%20\""}, "link"))), "[link](foo%20\\\" \"bar\")")

	// escapes list markers inside lists
	same("* 1\\. hi\n\n* x",
		doc(ul(li(p("1. hi")), li(p("x")))))

	// code block fence adjusts to content
	same("````\n```\ncode\n```\n````", doc(pre("```\ncode\n```")))

	// parses a code block ends with empty line
	originalText := "1\n"
	attrs := map[string]interface{}{"params": ""}
	content := []*model.Node{schema.Text(originalText)}
	mdText := DefaultSerializer.Serialize(
		doc(schema.Node("code_block", attrs, content)).Node)
	same(mdText, doc(schema.Node("code_block", attrs, content)))

	// doesn't create an empty text
	same("**foo**\\\nbar",
		doc(p(strong("foo"), br, "bar")))
}
