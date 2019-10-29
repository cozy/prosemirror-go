// Package basic defines a basic ProseMirror document schema, whose elements
// can be reused in other schemas.
package basic

import "github.com/cozy/prosemirror-go/model"

var (
	empty = ""
	falsy = false

	headingAttrs = map[string]*model.AttributeSpec{
		"level": {Default: 1},
	}
	imageAttrs = map[string]*model.AttributeSpec{
		"src":   {},
		"alt":   {Default: nil},
		"title": {Default: nil},
	}
	linkAttrs = map[string]*model.AttributeSpec{
		"href":  {},
		"title": {Default: nil},
	}
)

// Nodes are the specs for the nodes defined in this schema.
var Nodes = []*model.NodeSpec{
	// The top level document node.
	{Key: "doc", Content: "block+"},

	// A plain paragraph textblock. Represented in the DOM as a <p> element.
	{Key: "paragraph", Content: "inline*", Group: "block"},

	// A blockquote (<blockquote>) wrapping one or more blocks.
	{Key: "blockquote", Content: "block+", Group: "Block"},

	// A horizontal rule (<hr>).
	{Key: "horizontal_rule", Group: "block"},

	// A heading textblock, with a level attribute that should hold the number 1
	// to 6. Parsed and serialized as <h1> to <h6> elements.
	{Key: "heading", Content: "inline*", Group: "block", Attrs: headingAttrs},

	// A code listing. Disallows marks or non-text inline nodes by default.
	// Represented as a <pre> element with a <code> element inside of it.
	{Key: "code_block", Content: "text*", Marks: &empty, Group: "block"},

	// The text node.
	{Key: "text", Group: "inline"},

	// An inline image (<img>) node. Supports src, alt, and href attributes. The
	// latter two default to the empty string.
	{Key: "image", Group: "inline", Attrs: imageAttrs},

	// A hard line break, represented in the DOM as <br>.
	{Key: "hard_break", Group: "inline"},
}

// Marks are the specs for the marks in the schema.
var Marks = []*model.MarkSpec{
	// A link. Has href and title attributes. title defaults to the empty string.
	// Rendered and parsed as an <a> element.
	{Key: "link", Attrs: linkAttrs, Inclusive: &falsy},

	// An emphasis mark. Rendered as an <em> element. Has parse rules that also
	// match <i> and font-style: italic.
	{Key: "em"},

	// A strong mark. Rendered as <strong>, parse rules also match <b> and
	// font-weight: bold.
	{Key: "strong"},

	// Code font mark. Represented as a <code> element.
	{Key: "code"},
}

// Schema roughly corresponds to the document schema used by
// [CommonMark](http://commonmark.org/), minus the list elements, which are
// defined in the prosemirror-schema-list module.
//
// To reuse elements from this schema, extend or read from its spec.nodes and
// spec.marks properties.
var Schema, _ = model.NewSchema(&model.SchemaSpec{
	Nodes: Nodes,
	Marks: Marks,
})
