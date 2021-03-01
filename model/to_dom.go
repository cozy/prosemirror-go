package model

import (
	"fmt"
	"strconv"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ToDOM function type
type ToDOM = func(NodeOrMark) *html.Node

type NodeOrMark interface {
	GetAttrs([]string) []html.Attribute
}

func (n *Node) GetAttrs(attrs []string) []html.Attribute {
	result := []html.Attribute{}
	for key, value := range n.Attrs {
		result = addAttr(key, value, result)
	}
	return result
}

func (m *Mark) GetAttrs(selectedAttrs []string) []html.Attribute {
	result := []html.Attribute{}
	for key, value := range m.Attrs {
		for _, a := range selectedAttrs {
			if a == key {
				result = addAttr(key, value, result)
				break
			}
		}
	}
	return result
}

func addAttr(key string, value interface{}, attrs []html.Attribute) []html.Attribute {
	newAttr := html.Attribute{
		Key: key,
	}
	if attrInt, ok := value.(int); ok {
		newAttr.Val = strconv.Itoa(attrInt)
		return append(attrs, newAttr)
	} else {
		if attrString, ok := value.(string); ok {
			newAttr.Val = attrString
			return append(attrs, newAttr)
		}
	}
	return attrs
}

func defaultDOMGenerator(atom atom.Atom, attrs []string) ToDOM {
	return func(n NodeOrMark) *html.Node {
		htmlAttrs := n.GetAttrs(attrs)
		return &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom,
			Data:     atom.String(),
			Attr:     htmlAttrs,
		}
	}
}

func defaultCodeBlockDOMGenerator() ToDOM {
	return func(n NodeOrMark) *html.Node {
		outerNode := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Pre,
			Data:     "pre",
			Attr:     nil,
		}
		innerNode := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Code,
			Data:     "code",
			Attr:     nil,
		}
		outerNode.AppendChild(innerNode)
		return outerNode
	}
}

func defaultHeadingDOMGenerator() ToDOM {
	return func(n NodeOrMark) *html.Node {
		var dataAtom atom.Atom
		attrs := n.GetAttrs([]string{"level"})
		level := "1"
		for _, a := range attrs {
			if a.Key == "level" {
				level = a.Val
				break
			}
		}
		switch level {
		case "1":
			dataAtom = atom.H1
		case "2":
			dataAtom = atom.H2
		case "3":
			dataAtom = atom.H3
		case "4":
			dataAtom = atom.H3
		case "5":
			dataAtom = atom.H3
		case "6":
			dataAtom = atom.H3
		default:
			dataAtom = atom.H1
		}

		return &html.Node{
			Type:     html.ElementNode,
			DataAtom: dataAtom,
			Data:     "h" + level,
			Attr:     nil,
		}
	}
}

// Default ToDOM functions
var (
	defaultToDOM = map[string]ToDOM{
		"paragraph":       defaultDOMGenerator(atom.P, nil),
		"blockquote":      defaultDOMGenerator(atom.Blockquote, nil),
		"horizontal_rule": defaultDOMGenerator(atom.Hr, nil),
		"image":           defaultDOMGenerator(atom.Img, []string{"src"}),
		"hard_break":      defaultDOMGenerator(atom.Br, nil),
		"bullet_list":     defaultDOMGenerator(atom.Ul, nil),
		"ordered_list":    defaultDOMGenerator(atom.Ol, nil),
		"list_item":       defaultDOMGenerator(atom.Li, nil),
		"code_block":      defaultCodeBlockDOMGenerator(),
		"heading":         defaultHeadingDOMGenerator(),
	}
	defaultMarkToDOM = map[string]ToDOM{
		// link
		"em":     defaultDOMGenerator(atom.Em, nil),
		"strong": defaultDOMGenerator(atom.Strong, nil),
		"code":   defaultDOMGenerator(atom.Code, nil),
	}
)

// A DOM serializer knows how to convert ProseMirror nodes and
// marks of various types to DOM nodes.
type DOMSerializer struct {
	// Create a serializer. `Nodes` should map node names to functions
	// that take a node and return a description of the corresponding
	// DOM. `Marks` does the same for mark names, but also gets an
	// argument that tells it whether the mark's content is block or
	// inline content (for typical use, it'll always be inline). A mark
	// serializer may be `null` to indicate that marks of that type
	// should not be serialized.

	// The node serialization functions.
	Nodes map[string]ToDOM

	// The mark serialization functions.
	Marks map[string]ToDOM
}

// Helper function to add default ToDOM functions to schema
func AddDefaultToDOM(schema *Schema) *Schema {
	result := schema
	for i, n := range result.Nodes {
		if n.Spec.ToDOM == nil {
			if defaultToDOM, ok := defaultToDOM[n.Name]; ok {
				result.Nodes[i].Spec.ToDOM = defaultToDOM
			}
		}
	}
	for i, m := range result.Marks {
		if m.Spec.ToDOM == nil {
			if defaultToDOM, ok := defaultMarkToDOM[m.Name]; ok {
				result.Marks[i].Spec.ToDOM = defaultToDOM
			}
		}
	}
	return result
}

// Build a serializer using the properties in a schema's node and
// mark specs.
func DOMSerializerFromSchema(schema *Schema) *DOMSerializer {
	return &DOMSerializer{
		Nodes: nodesFromSchema(schema),
		Marks: marksFromSchema(schema),
	}
}

// Helper function
func (d *DOMSerializer) hasMark(markName string) bool {
	for key := range d.Marks {
		if key == markName {
			return true
		}
	}
	return false
}

// Serialize the content of this fragment to HTML.
func (d *DOMSerializer) SerializeFragment(fragment *Fragment, options interface{}, target *html.Node) *html.Node {
	if target == nil {
		target = &html.Node{
			Type: html.DocumentNode,
		}
	}
	type activeMark struct {
		mark *Mark
		top  *html.Node
	}
	var active []activeMark
	top := target
	fragment.ForEach(func(node *Node, offset, index int) {

		fmt.Printf("  Node name: %s\n", node.Type.Name)
		for key, val := range node.Attrs {
			fmt.Printf("  Node attributes: %s:%s\n", key, val)
		}
		if active != nil || len(node.Marks) > 0 {
			keep, rendered := 0, 0
			for keep < len(active) && rendered < len(node.Marks) {
				next := node.Marks[rendered]
				if !d.hasMark(next.Type.Name) {
					rendered++
					continue
				}
				if !next.Eq(active[keep].mark) || (next.Type.Spec.Spanning != nil && !*next.Type.Spec.Spanning) {
					break
				}
				keep++
				rendered++
			}
			for keep < len(active) {
				n := len(active)
				top, active = active[n-1].top, active[:n-1]
			}
			for rendered < len(node.Marks) {
				add := node.Marks[rendered]
				rendered++
				markDOM := d.serializeMark(add, node.IsInline())
				if markDOM != nil {
					active = append(active, activeMark{mark: add, top: top})
					top.AppendChild(markDOM)
					top = markDOM
				}
			}

		}
		child := d.SerializeNode(node)
		if child != nil {
			top.AppendChild(child)
		}
	})
	return target
}

func (d *DOMSerializer) serializeMark(mark *Mark, inline bool) *html.Node {
	toDOM := d.Marks[mark.Type.Name]
	return toDOM(mark)

	//return toDOM && DOMSerializer.renderSpec(doc(options), toDOM(mark, inline))
}

// Serialize this node to a DOM node. This can be useful when you
// need to serialize a part of a document, as opposed to the whole
// document. To serialize a whole document, use serializeFragment()
func (d *DOMSerializer) SerializeNode(node *Node) *html.Node {
	domFn := d.Nodes[node.Type.Name]
	if domFn != nil {
		fmt.Printf("  Type of node: %s\n", node.Type.Name)
		topNode := domFn(node)
		contentNode := topNode
		for contentNode.FirstChild != nil {
			contentNode = contentNode.FirstChild
		}
		d.SerializeFragment(node.Content, nil, contentNode)
		return topNode
	}
	return nil
}

// Gather the serializers in a schema's node specs into an object.
// This can be useful as a base to build a custom serializer from.
func nodesFromSchema(schema *Schema) (result map[string]ToDOM) {
	result = make(map[string]ToDOM)
	for _, n := range schema.Nodes {
		result[n.Name] = n.Spec.ToDOM
	}
	if textToDOM, ok := result["text"]; ok && textToDOM == nil {
		result["text"] = func(n NodeOrMark) *html.Node {
			node, _ := n.(*Node)
			return &html.Node{
				Type: html.TextNode,
				Data: *node.Text,
			}
		}
	}
	return result
}

// Gather the serializers in a schema's mark specs into an object.
func marksFromSchema(schema *Schema) (result map[string]ToDOM) {
	result = make(map[string]ToDOM)
	for _, m := range schema.Marks {
		result[m.Name] = m.Spec.ToDOM
	}
	return result
}
