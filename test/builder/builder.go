package builder

import (
	"github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/schema/basic"
	"github.com/cozy/prosemirror-go/schema/list"
)

type Spec map[string]interface{}
type NodeBuilder func(args ...interface{}) *model.Node
type MarkBuilder func(args ...interface{}) *model.Mark

func takeAttrs(attrs map[string]interface{}, args []interface{}) map[string]interface{} {
	return attrs // TODO
}

func block(typ *model.NodeType, attrs map[string]interface{}) NodeBuilder {
	return func(args ...interface{}) *model.Node {
		// TODO
		node, err := typ.Create(takeAttrs(attrs, args), nil, nil)
		if err != nil {
			panic(err)
		}
		return node
	}
	// TODO if (type.isLeaf) try { result.flat = [type.create(attrs)] } catch(_) {}
}

// Create a builder function for marks.
func mark(typ *model.MarkType, attrs map[string]interface{}) MarkBuilder {
	return func(args ...interface{}) *model.Mark {
		// TODO
		return typ.Create(takeAttrs(attrs, args))
	}
}

func Builders(schema *model.Schema, names map[string]Spec) map[string]interface{} {
	result := map[string]interface{}{"schema": schema}
	for name, typ := range schema.Nodes {
		result[name] = block(typ, nil)
	}
	for name, typ := range schema.Marks {
		result[name] = mark(typ, nil)
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
	Schema     = out["schema"].(*model.Schema)
	Doc        = out["doc"].(NodeBuilder)
	P          = out["p"].(NodeBuilder)
	Blockquote = out["blockquote"].(NodeBuilder)
	Pre        = out["pre"].(NodeBuilder)
	H1         = out["h1"].(NodeBuilder)
	H2         = out["h2"].(NodeBuilder)
	H3         = out["h3"].(NodeBuilder)
	Li         = out["li"].(NodeBuilder)
	Ul         = out["ul"].(NodeBuilder)
	Ol         = out["ol"].(NodeBuilder)
	Br         = out["br"].(NodeBuilder)
	Img        = out["img"].(NodeBuilder)
	Hr         = out["hr"].(NodeBuilder)
	A          = out["a"].(MarkBuilder)
	Em         = out["em"].(MarkBuilder)
	Strong     = out["strong"].(MarkBuilder)
	Code       = out["code"].(MarkBuilder)
)
