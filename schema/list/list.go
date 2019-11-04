// Package list exports list-related schema elements and commands. The commands
// assume lists to be nestable, with the restriction that the first child of a
// list item is a plain paragraph.
package list

import "github.com/cozy/prosemirror-go/model"

var (
	// An ordered list node spec. Has a single attribute, order, which
	// determines the number at which the list starts counting, and defaults to
	// 1. Represented as an <ol> element.
	orderedList = model.NodeSpec{
		Key: "ordered_list",
		Attrs: map[string]*model.AttributeSpec{
			"order": {Default: 1.0},
		},
	}

	// A bullet list node spec, represented in the DOM as <ul>.
	bulletList = model.NodeSpec{
		Key: "bullet_list",
	}

	// A list item (<li>) spec.
	listItem = model.NodeSpec{
		Key: "list_item",
	}
)

func add(obj, props model.NodeSpec) *model.NodeSpec {
	if props.Content != "" {
		obj.Content = props.Content
	}
	if props.Group != "" {
		obj.Group = props.Group
	}
	return &obj
}

// AddListNodes is a onvenience function for adding list-related node types to
// a map specifying the nodes for a schema. Adds orderedList as "ordered_list",
// bulletList as "bullet_list", and listItem as "list_item".
//
// itemContent determines the content expression for the list items. If you
// want the commands defined in this module to apply to your list structure, it
// should have a shape like "paragraph block*" or "paragraph (ordered_list |
// bullet_list)*". listGroup can be given to assign a group name to the list
// node types, for example "block".
func AddListNodes(nodes []*model.NodeSpec, itemContent, listGroup string) []*model.NodeSpec {
	return append(
		nodes,
		add(orderedList, model.NodeSpec{Content: "list_item+", Group: listGroup}),
		add(bulletList, model.NodeSpec{Content: "list_item+", Group: listGroup}),
		add(listItem, model.NodeSpec{Content: itemContent}),
	)
}
