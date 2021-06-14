package model

import (
	"fmt"

	"github.com/dstotijn/go-notion"
)

// ToDOM function type
type ToNotionBlock = func(*Node) *notion.Block

type NotionSerializer struct {
	// The node serialization functions.
	Nodes map[string]ToNotionBlock

	// The mark serialization functions.
	Marks map[string]ToNotionBlock
}

func CreatePageContent(node *Node, schema *Schema) []notion.Block {
	s := AddDefaultToNotion(schema)
	serializer := NotionSerializerFromSchema(s)
	output := serializer.SerializePage(node.Content)
	return output
}

/*
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
	} else if attrString, ok := value.(string); ok {
		newAttr.Val = attrString
		return append(attrs, newAttr)
	} else if attrBool, ok := value.(bool); ok {
		newAttr.Val = strconv.FormatBool(attrBool)
		return append(attrs, newAttr)
	}
	return attrs
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


*/

func defaultNotionGenerator(blocktype notion.BlockType, attrs []string) ToNotionBlock {
	return func(n *Node) *notion.Block {
		//htmlAttrs := n.GetAttrs(attrs)
		return &notion.Block{
			Type: blocktype,
			//Type:     html.ElementNode,
			//DataAtom: atom,
			//Data:     atom.String(),
			//Attr:     htmlAttrs,
		}
	}
}

func defaultParagraphBlockGenerator() ToNotionBlock {
	return func(n *Node) *notion.Block {
		return &notion.Block{
			Type:      notion.BlockTypeParagraph,
			Paragraph: createParagraphBlock(n),
			//Type:     html.ElementNode,
			//DataAtom: atom,
			//Data:     atom.String(),
			//Attr:     htmlAttrs,
		}
	}
}

func createParagraphBlock(n *Node) *notion.RichTextBlock {
	result := &notion.RichTextBlock{
		Text: []notion.RichText{},
	}
	n.ForEach(func(node *Node, offset, index int) {
		text := ""
		annotations := &notion.Annotations{}
		hasAnnotation := false
		if node.Type.Name == "text" {
			text = text + *node.Text
		}
		if node.Type.Name == "hard_break" {
			text = text + "/n"
		}
		for _, m := range node.Marks {
			if m.Type.Name == "em" {
				annotations.Bold = true
				hasAnnotation = true
			}
			if m.Type.Name == "strong" {
				annotations.Bold = true
				hasAnnotation = true
			}

		}
		nextRichText := &notion.RichText{
			Type:      notion.RichTextTypeText,
			PlainText: text,
			Text: &notion.Text{
				Content: text,
			},
		}
		if hasAnnotation {
			nextRichText.Annotations = annotations
		}
		result.Text = append(result.Text, *nextRichText)
	})
	return result
}

// Build a serializer using the properties in a schema's node and
// mark specs.
func NotionSerializerFromSchema(schema *Schema) *NotionSerializer {
	return &NotionSerializer{
		Nodes: notionNodesFromSchema(schema),
		//Marks: marksFromSchema(schema),
	}
}

// Default ToDOM functions
var (
	defaultToNotion = map[string]ToNotionBlock{
		"paragraph": defaultParagraphBlockGenerator(),
		//"blockquote":      defaultNotionGenerator(atom.Blockquote, nil),
		//"horizontal_rule": defaultNotionGenerator(atom.Hr, nil),
		//"image":           defaultNotionGenerator(atom.Img, []string{"src"}),
		//"hard_break":      defaultNotionGenerator(atom.Br, nil),
		//"bullet_list":     defaultNotionGenerator(atom.Ul, nil),
		//"ordered_list":    defaultNotionGenerator(atom.Ol, nil),
		//"list_item":       defaultNotionGenerator(atom.Li, nil),
		//"code_block":      defaultCodeBlockDOMGenerator(),
		//"heading":         defaultHeadingDOMGenerator(),
	}
	//defaultMarkToDOM = map[string]ToDOM{
	// link
	//"em":     defaultDOMGenerator(atom.Em, nil),
	//"strong": defaultDOMGenerator(atom.Strong, nil),
	//"code":   defaultDOMGenerator(atom.Code, nil),
	//}
)

// Helper function to add default ToNotion functions to schema
func AddDefaultToNotion(schema *Schema) *Schema {
	result := schema
	for i, n := range result.Nodes {
		if n.Spec.ToNotion == nil {
			if defaultToNotion, ok := defaultToNotion[n.Name]; ok {
				result.Nodes[i].Spec.ToNotion = defaultToNotion
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

// Helper function
func (n *NotionSerializer) hasMark(markName string) bool {
	for key := range n.Marks {
		if key == markName {
			return true
		}
	}
	return false
}

// Serialize the content of this fragment to HTML.
func (n *NotionSerializer) SerializePage(fragment *Fragment) []notion.Block {

	//type activeMark struct {
	//mark *Mark
	//top  *html.Node
	//}
	var result []notion.Block
	fragment.ForEach(func(node *Node, offset, index int) {

		fmt.Printf("  Node name: %s\n", node.Type.Name)
		for key, val := range node.Attrs {
			fmt.Printf("  Node attributes: %s:%s\n", key, val)
		}
		nextBlock := n.SerializeNode(node)
		if nextBlock != nil {
			result = append(result, *nextBlock)
		}

		//if child != nil {
		//top.AppendChild(child)
		//top.HasChildren = true
		//}
	})
	return result
}

//func (n *NotionSerializer) serializeMark(mark *Mark, inline bool) *html.Node {
//toDOM := d.Marks[mark.Type.Name]
//return toDOM(mark)

//return toDOM && DOMSerializer.renderSpec(doc(options), toDOM(mark, inline))
//}

// Serialize this node to a DOM node. This can be useful when you
// need to serialize a part of a document, as opposed to the whole
// document. To serialize a whole document, use serializeFragment()
func (n *NotionSerializer) SerializeNode(node *Node) *notion.Block {
	notionFn := n.Nodes[node.Type.Name]
	if notionFn != nil {
		fmt.Printf("  Type of node: %s\n", node.Type.Name)
		topNode := notionFn(node)
		//contentNode := topNode
		//node.ForEach(func(node *Node, offset, index int) {
		//contentNode.HasChildren = true
		//})
		//for contentNode.LastChild != nil {
		//	contentNode = contentNode.LastChild
		//}
		//n.SerializeFragment(node.Content, nil, contentNode)
		return topNode
	}
	return nil
}

func notionNodesFromSchema(schema *Schema) (result map[string]ToNotionBlock) {
	result = make(map[string]ToNotionBlock)
	for _, n := range schema.Nodes {
		result[n.Name] = n.Spec.ToNotion
	}
	if textToNotion, ok := result["text"]; ok && textToNotion == nil {
		result["text"] = func(node *Node) *notion.Block {
			return &notion.Block{
				Type: notion.BlockTypeParagraph,
				Paragraph: &notion.RichTextBlock{
					Text: []notion.RichText{
						notion.RichText{
							PlainText: *node.Text,
						},
					},
				},
			}
		}
	}
	return result
}

/*
// Gather the serializers in a schema's node specs into an object.
// This can be useful as a base to build a custom serializer from.

// Gather the serializers in a schema's mark specs into an object.
func marksFromSchema(schema *Schema) (result map[string]ToDOM) {
	result = make(map[string]ToDOM)
	for _, m := range schema.Marks {
		result[m.Name] = m.Spec.ToDOM
	}
	return result
}
*/
