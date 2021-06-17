package model_test

import (
	"fmt"
	"testing"

	"github.com/dstotijn/go-notion"
	. "github.com/shodgson/prosemirror-go/model"
	"github.com/shodgson/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func notionTest(t *testing.T, doc builder.NodeWithTag, notionJSON string, serializer *NotionSerializer, msg string) {
	fmt.Println("Testing: " + msg)
	output := serializer.SerializePage(doc.Content)
	outputValues := []notion.Block{}
	for _, o := range output {
		outputValues = append(outputValues, *o)
	}
	pageParams := &notion.CreatePageParams{
		Children: outputValues,
	}
	result, err := pageParams.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, notionJSON, string(result), msg)
}

func newPage(children string) string {
	return fmt.Sprintf(`{"parent":{},"properties":null,"children":[%s]}`, children)

}

func TestNotionParser(t *testing.T) {

	//schema := AddDefaultToNotion(builder.Schema)
	//serializer := NotionSerializerFromSchema(schema)

	/*
		notionTest(t,
			doc(p("hello")),
			newPage(`{"object":"block","type":"paragraph","paragraph":{"text":[{"type":"text","plain_text":"hello","text":{"content":"hello"}}]}}`),
			serializer,
			"Should represent simple node")
	*/

	/*
		notionTest(t,
			doc(p("hi", br, "there")),
			//newPage(`{"object":"block","type":"paragraph","paragraph":{"text":[{"type":"text","plain_text":"hi/nthere","text":{"content":"hi/nthere"}}]}}`),
			newPage(`{"object":"block","type":"paragraph","paragraph":{"text":[{"type":"text","plain_text":"hi","text":{"content":"hi"}},{"type":"text","plain_text":"/n","text":{"content":"/n"}},{"type":"text","plain_text":"there","text":{"content":"there"}}]}}`),
			serializer,
			"Should represent a line break")
	*/

	/*

		//test(t,
		//	doc(p("hi", imageWithAttrs("x", "img.png"), "there")),
		//	[]string{`<p>hi<img src="img.png" alt="x"/>there</p>`, `<p>hi<img alt="x" src="img.png"/>there</p>`},
		//	serializer,
		//	"Should represent an image")

	*/

	/*
		notionTest(t,
			doc(p(em("emphasis"))),
			newPage(`{"object":"block","type":"paragraph","paragraph":{"text":[{"type":"text","annotations":{"bold":true},"plain_text":"emphasis","text":{"content":"emphasis"}}]}}`),
			serializer,
			"Should represent simple marks")

	*/

	/*
		//Currently failing. Outcome is still valid, but the styles are not joined
		//notionTest(t,
		//doc(p("one", strong("two", em("three")), em("four"), "five")),
		// TODO: newPage(`{"object":"block","type":"paragraph","paragraph":{"text":[{"type":"text","plain_text":"one","text":{"content":"one"}}]}}`),
		//serializer,
		//"Should join styles")

		// Currently failing. Nested links not supported in current builder
		//test(t,
		//doc(p("a ", link("foo", "big ", link("bar", "nested"), " link"))),
		//"<p>a <a href=\"foo\">big </a><a href=\"bar\">nested</a><a href=\"foo\"> link</a></p>",
		//serializer,
		//"Can represent links")
		//TODO: test links

	*/

	/*
		notionTest(t,
			doc(ul(li(p("one")), li(p("two")), li(p("three", strong("!")))), p("after")),
			"<ul><li><p>one</p></li><li><p>two</p></li><li><p>three<strong>!</strong></p></li></ul><p>after</p>",
			serializer,
			"Should represent an unordered list")

		/*
			test(t,
				doc(ol(li(p("one")), li(p("two")), li(p("three", strong("!")))), p("after")),
				"<ol><li><p>one</p></li><li><p>two</p></li><li><p>three<strong>!</strong></p></li></ol><p>after</p>",
				serializer,
				"Should represent an ordered list")

			test(t,
				doc(blockquote(p("hello"), p("bye"))),
				"<blockquote><p>hello</p><p>bye</p></blockquote>",
				serializer,
				"Should represent a blockquote")

			test(t,
				doc(blockquote(blockquote(blockquote(p("he said"))), p("i said"))),
				"<blockquote><blockquote><blockquote><p>he said</p></blockquote></blockquote><p>i said</p></blockquote>",
				serializer,
				"Should represent a nested blockquote")

			test(t,
				doc(h1("one"), h2("two"), p("text")),
				"<h1>one</h1><h2>two</h2><p>text</p>",
				serializer,
				"Should represent headings")

			// Test modified from dom_test.js to include different order in tags
			test(t,
				doc(p("text and ", code("code that is ", em("emphasized"), "..."))),
				[]string{"<p>text and <code>code that is </code><em><code>emphasized</code></em><code>...</code></p>",
					"<p>text and <code>code that is </code><code><em>emphasized</em></code><code>...</code></p>"},
				serializer,
				"Should represent inline code")

			// Test modified from dom_test.js to include different order in tags
			test(t,
				doc(blockquote(pre("some code")), p("and")),
				"<blockquote><pre><code>some code</code></pre></blockquote><p>and</p>",
				serializer,
				"Should represent a code block")

			//"<p><em>hi</em><em><br/></em><em>x</em></p>
			test(t,
				doc(p(em("hi", br, "x"))),
				[]string{"<p><em>hi<br>x</em></p>",
					"<p><em>hi<br/>x</em></p>"},
				serializer,
				"Supports leaf nodes in marks")

			test(t,
				doc(p("\u00a0 \u00a0hello\u00a0")),
				"<p>\u00a0 \u00a0hello\u00a0</p>",
				serializer,
				"Should not collapse non-breaking spaces")
	*/

	return
}

/*
func TestMarksOnBlockNodes(t *testing.T) {
	commentToDOM := func(n NodeOrMark) *html.Node {
		return &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Div,
			Data:     "div",
			Attr: []html.Attribute{
				{Key: "class", Val: "comment"},
			},
		}
	}
	commentSpec := &MarkSpec{Key: "comment", ToDOM: commentToDOM}
	commentSchema, err := NewSchema(&SchemaSpec{
		Nodes: builder.Schema.Spec.Nodes,
		Marks: append(builder.Schema.Spec.Marks, commentSpec),
	})
	assert.NoError(t, err)

	out := builder.Builders(commentSchema, nil)
	bComment := out["comment"].(builder.MarkBuilder)
	bParagraph := out["paragraph"].(builder.NodeBuilder)
	bDoc := out["doc"].(builder.NodeBuilder)
	bStrong := out["strong"].(builder.MarkBuilder)

	commentSerializer := DOMSerializerFromSchema(commentSchema)

	test(t,
		bDoc(bParagraph("one"), bComment(bParagraph("two"), bParagraph(bStrong("three"))), bParagraph("four")),
		"<p>one</p><div class=\"comment\"><p>two</p><p><strong>three</strong></p></div><p>four</p>",
		commentSerializer,
		"Should parse marks on block nodes")

	return
}

func imageWithAttrs(alt string, src string) *Node {
	attrs := map[string]interface{}{"alt": alt, "src": src}
	image, err := schema.Node("image", attrs)
	if err != nil {
		return nil
	}
	return image
}
*/
