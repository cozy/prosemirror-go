// Package builder is a module used to write tests for ProseMirror. ProseMirror
// is a well-behaved rich semantic content editor based on contentEditable,
// with support for collaborative editing and custom document schemas.
//
// This module provides helpers for building ProseMirror documents for tests.
// It's main file exports a basic schema with list support, and a number of
// functions, whose name mostly follows the corresponding HTML tag, to create
// nodes and marks in this schema. The prosemirror-test-builder/dist/build
// module exports a function that you can use to create such helpers for your
// own schema.
//
// Node builder functions optionally take an attribute object as their first
// argument, followed by zero or more child nodes, and return a node with those
// attributes and children. Children should be either strings (for text nodes),
// existing nodes, or the result of calling a mark builder. For leaf nodes, you
// may also pass the builder function itself, without calling it. Mark builder
// functions work similarly, but return an object representing a set of nodes
// rather than a single node.
//
// These builders help specifying and retrieving positions in the documents
// that you created (to avoid needing to count tokens when writing tests).
// Inside of strings passed as child nodes, angle-brackets <name> syntax can be
// used to place a tag called name at that position. The angle-bracketed part
// will not appear in the result node, but is stored in the node's tag
// property, which is an object mapping tag names to position integers. A
// string which is only a tag or set of tags may appear everywhere, even in
// places where text nodes aren't allowed.
//
// So if you've imported doc and p from this module, the expression
// doc(p("foo<a>")) will return a document containing a single paragraph, and
// its .tag.a will hold the number 4 (the position at the end of the
// paragraph).
package builder

import (
	"fmt"

	"github.com/shodgson/prosemirror-go/model"
	"github.com/shodgson/prosemirror-go/schema/basic"
	"github.com/shodgson/prosemirror-go/schema/list"
)

// Spec can be used to add custom builders—if given, it should be an object
// mapping names to attribute objects, which may contain a nodeType or markType
// property to specify which node or mark the builder by this name should
// create.
type Spec map[string]interface{}

// Result is what is returned by a MarkBuilder.
type Result struct {
	Nodes []*model.Node
	Flat  []*model.Node
	Tag   map[string]int
}

// NodeWithTag is what is returned by a NodeBuilder.
type NodeWithTag struct {
	*model.Node
	Tag map[string]int
}

// NodeBuilder functions optionally take an attribute object as their first
// argument, followed by zero or more child nodes, and return a node with those
// attributes and children. Children should be either strings (for text nodes),
// existing nodes, or the result of calling a mark builder. For leaf nodes, you
// may also pass the builder function itself, without calling it.
type NodeBuilder func(args ...interface{}) NodeWithTag

// MarkBuilder functions work similarly to NodeBuilder, but return an object
// representing a set of nodes rather than a single node.
type MarkBuilder func(args ...interface{}) Result

type nodeMapper func(n *model.Node) *model.Node

func id(n *model.Node) *model.Node { return n }

func flatten(schema *model.Schema, children []interface{}, f nodeMapper) Result {
	result := []*model.Node{}
	pos := 0
	tag := map[string]int{}

	for _, child := range children {
		switch child := child.(type) {
		case NodeWithTag:
			for id, val := range child.Tag {
				extra := 0
				if !child.IsText() {
					extra = 1
				}
				tag[id] = val + extra + pos
			}
		case Result:
			extra := 0
			if len(child.Flat) == 0 {
				extra = 1
			}
			for id, val := range child.Tag {
				tag[id] = val + extra + pos
			}
		}

		switch child := child.(type) {
		case string:
			at := 0
			out := ""
			for i, c := range child { // i == m.index
				if c != '<' {
					continue
				}
				space := false
				for j, c := range child[i:] { // j == m[0].length
					if c == ' ' {
						space = true
						break
					}
					if c == '>' {
						out += child[at:i]
						pos += i - at
						at = i + j + 1
						tag[child[i+1:i+j]] = pos
						break
					}
				}
				if space {
					break
				}
			}
			out += child[at:]
			pos += len(child) - at
			if len(out) > 0 {
				result = append(result, f(schema.Text(out)))
			}
		case NodeWithTag:
			node := f(child.Node)
			pos += node.NodeSize()
			result = append(result, node)
		case Result:
			for j := range child.Flat {
				node := f(child.Flat[j])
				pos += node.NodeSize()
				result = append(result, node)
			}
		case NodeBuilder:
			node := f(child().Node)
			pos += node.NodeSize()
			result = append(result, node)
		case *model.Node:
			node := f(child)
			pos += node.NodeSize()
			result = append(result, node)
		default:
			fmt.Printf("Unknown test type: %T (%v)\n", child, child)
		}
	}

	return Result{Nodes: result, Tag: tag}
}

func takeAttrs(attrs map[string]interface{}, args []interface{}) (map[string]interface{}, []interface{}) {
	if len(args) == 0 {
		return attrs, args
	}
	a0 := args[0]
	switch a0 := a0.(type) {
	case string, *model.Node, NodeWithTag, Result, NodeBuilder:
		return attrs, args

	case map[string]interface{}:
		args = args[1:]
		if len(attrs) == 0 {
			return a0, args
		}
		if len(a0) == 0 {
			return attrs, args
		}
		result := map[string]interface{}{}
		for k, v := range attrs {
			result[k] = v
		}
		for k, v := range a0 {
			result[k] = v
		}
		return result, args
	}
	panic(fmt.Errorf("Unsupported type %T for takeAttrs (%v)", a0, a0))
}

func block(typ *model.NodeType, attrs map[string]interface{}) NodeBuilder {
	return func(args ...interface{}) NodeWithTag {
		myAttrs, myArgs := takeAttrs(attrs, args)
		//fmt.Println(myAttrs)
		//fmt.Println(myArgs)
		result := flatten(typ.Schema, myArgs, id)
		node, err := typ.Create(myAttrs, result.Nodes, nil)
		if err != nil {
			panic(err)
		}
		nt := NodeWithTag{Node: node}
		if len(result.Tag) > 0 {
			nt.Tag = result.Tag
		}
		return nt
	}
}

// Create a builder function for marks.
func mark(typ *model.MarkType, attrs map[string]interface{}) MarkBuilder {
	return func(args ...interface{}) Result {
		myAttrs, myArgs := takeAttrs(attrs, args)
		mark := typ.Create(myAttrs)
		f := func(n *model.Node) *model.Node {
			if mark.Type.IsInSet(n.Marks) != nil {
				return n
			}
			return n.Mark(mark.AddToSet(n.Marks))
		}
		result := flatten(typ.Schema, myArgs, f)
		return Result{
			Flat: result.Nodes,
			Tag:  result.Tag,
		}
	}
}

// Builders can be called with a schema and an optional object of
// renamed/configured builders to create a object of builders for a custom
// schema. It will return an object with a schema property and one builder for
// each node and mark in the schema. The second argument can be used to add
// custom builders—if given, it should be an object mapping names to attribute
// objects, which may contain a nodeType or markType property to specify which
// node or mark the builder by this name should create.
func Builders(schema *model.Schema, names map[string]Spec) map[string]interface{} {
	result := map[string]interface{}{"schema": schema}
	for _, typ := range schema.Nodes {
		result[typ.Name] = block(typ, nil)
	}
	for _, typ := range schema.Marks {
		result[typ.Name] = mark(typ, nil)
	}

	if len(names) > 0 {
		for name, value := range names {
			typeName, ok := value["nodeType"].(string)
			if !ok {
				typeName, ok = value["markType"].(string)
			}
			if !ok {
				typeName = name
			}
			if typ, err := schema.NodeType(typeName); err == nil {
				result[name] = block(typ, value)
			} else if typ, err := schema.MarkType(typeName); err == nil {
				result[name] = mark(typ, value)
			}
		}
	}

	return result
}

var testSchema, _ = model.NewSchema(&model.SchemaSpec{
	Nodes: list.AddListNodes(basic.Schema.Spec.Nodes, "paragraph block*", "block"),
	Marks: basic.Schema.Spec.Marks,
})

var out = Builders(testSchema, map[string]Spec{
	"p":   {"nodeType": "paragraph"},
	"pre": {"nodeType": "code_block"},
	"h1":  {"nodeType": "heading", "level": 1},
	"h2":  {"nodeType": "heading", "level": 2},
	"h3":  {"nodeType": "heading", "level": 3},
	"li":  {"nodeType": "list_item"},
	"ul":  {"nodeType": "bullet_list"},
	"ol":  {"nodeType": "ordered_list"},
	"br":  {"nodeType": "hard_break"},
	"img": {"nodeType": "image", "src": "img.png"},
	"hr":  {"nodeType": "horizontal_rule"},
	"a":   {"markType": "link", "href": "foo"},
})

var (
	// Schema is the test schema itself.
	Schema = out["schema"].(*model.Schema)
	// Doc is a builder for top type of Schema (the document).
	Doc = out["doc"].(NodeBuilder)
	// P is a builder for paragraph nodes.
	P = out["p"].(NodeBuilder)
	// Blockquote is a builder for blockquote nodes.
	Blockquote = out["blockquote"].(NodeBuilder)
	// Pre is a builder for code block nodes.
	Pre = out["pre"].(NodeBuilder)
	// H1 is a builder for heading block nodes with the level attribute defaulting to 1.
	H1 = out["h1"].(NodeBuilder)
	// H2 is a builder for heading block nodes with the level attribute defaulting to 2.
	H2 = out["h2"].(NodeBuilder)
	// H3 is a builder for heading block nodes with the level attribute defaulting to 3.
	H3 = out["h3"].(NodeBuilder)
	// Li is a builder for list item nodes.
	Li = out["li"].(NodeBuilder)
	// Ul is a builder for bullet list nodes.
	Ul = out["ul"].(NodeBuilder)
	// Ol is a builder for ordered list nodes.
	Ol = out["ol"].(NodeBuilder)
	// Br is a builder for hard break nodes.
	Br = out["br"].(NodeBuilder)
	// Img is a builder for image nodes, with the src attribute defaulting to "img.png".
	Img = out["img"].(NodeBuilder)
	// Hr is a builder for horizontal rule nodes.
	Hr = out["hr"].(NodeBuilder)
	// A is a builder for link marks.
	A = out["a"].(MarkBuilder)
	// Em is a builder for em marks.
	Em = out["em"].(MarkBuilder)
	// Strong is a builder for strong marks.
	Strong = out["strong"].(MarkBuilder)
	// Code is a builder for code marks.
	Code = out["code"].(MarkBuilder)
)
