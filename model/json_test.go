package model_test

import (
	"encoding/json"
	"testing"

	"github.com/cozy/prosemirror-go/model"
	"github.com/stretchr/testify/assert"
)

func TestJSONNode(t *testing.T) {
	jsonSpec := `
{
  "nodes": [
    ["doc", { "content": "block+" }],
    ["paragraph", { "content": "inline*", "group": "block" }],
    ["blockquote", { "content": "block+", "group": "block" }],
    ["horizontal_rule", { "group": "block" }],
    [
  	"heading",
  	{
  	  "content": "inline*",
  	  "group": "block",
  	  "attrs": { "level": { "default": 1 } }
  	}
    ],
    ["code_block", { "content": "text*", "marks": "", "group": "block" }],
    ["text", { "group": "inline" }],
    [
  	"image",
  	{
  	  "group": "inline",
  	  "inline": true,
  	  "attrs": { "alt": {}, "src": {}, "title": {} }
  	}
    ],
    ["hard_break", { "group": "inline", "inline": true }],
    [
  	"ordered_list",
  	{
  	  "content": "list_item+",
  	  "group": "block",
  	  "attrs": { "order": { "default": 1 } }
  	}
    ],
    ["bullet_list", { "content": "list_item+", "group": "block" }],
    ["list_item", { "content": "paragraph block*" }]
  ],
  "marks": [
    ["link", { "attrs": { "href": {}, "title": {} }, "inclusive": false }],
    ["em", {}],
    ["strong", {}],
    ["code", {}]
  ],
  "topNode": "doc"
}`

	var spec model.SchemaSpec
	err := json.Unmarshal([]byte(jsonSpec), &spec)
	assert.NoError(t, err)
	schema, err := model.NewSchema(&spec)
	assert.NoError(t, err)
	typ, err := schema.NodeType(schema.Spec.TopNode)
	assert.NoError(t, err)
	node, err := typ.CreateAndFill()
	assert.NoError(t, err)
	assert.NotNil(t, node)
	result, err := json.Marshal(node.ToJSON())
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
