// Package model implements ProseMirror's document model, along with the
// mechanisms needed to support schemas.
package model

import (
	"errors"
	"fmt"
	"strings"
)

// Node types are objects allocated once per Schema and used to tag Node
// instances. They contain information about the node type, such as its name
// and what kind of node it represents.
type NodeType struct {
	// The name the node type has in this schema.
	Name string
	// A link back to the `Schema` the node type belongs to.
	Schema *Schema
	// The spec that this type is based on
	Spec  *NodeSpec
	Attrs map[string]interface{}
	// TODO
}

func NewNodeType(name string, schema *Schema, spec *NodeSpec) *NodeType {
	return &NodeType{
		Name:   name,
		Schema: schema,
		Spec:   spec,
		// TODO
	}
}

// True if this is the text node type.
func (nt *NodeType) IsText() bool {
	return nt.Name == "text"
}

func (nt *NodeType) computeAttrs(attrs map[string]interface{}) map[string]interface{} {
	return attrs // TODO
}

// Create a Node of this type. The given attributes are checked and defaulted
// (you can pass null to use the type's defaults entirely, if no required
// attributes exist). content may be a Fragment, a node, an array of nodes, or
// null. Similarly marks may be null to default to the empty set of marks.
func (nt *NodeType) Create(attrs map[string]interface{}, content interface{}, marks []*Mark) (*Node, error) {
	if nt.IsText() {
		return nil, errors.New("NodeType.create can't construct text nodes")
	}
	fragment, err := FragmentFrom(content)
	if err != nil {
		return nil, err
	}
	return NewNode(nt, nt.computeAttrs(attrs), fragment, MarkSetFrom(marks)), nil
}

func compileNodeType(nodes []*NodeSpec, schema *Schema) (map[string]*NodeType, error) {
	result := make(map[string]*NodeType)
	for _, n := range nodes {
		result[n.Key] = NewNodeType(n.Key, schema, n)
	}
	topType := schema.Spec.TopNode
	if _, ok := result[topType]; !ok {
		return nil, fmt.Errorf("The schema is missing its top node type (%s)", topType)
	}
	txt, ok := result["text"]
	if !ok {
		return nil, errors.New("Every schema needs a 'text' type")
	}
	if len(txt.Attrs) > 0 {
		return nil, errors.New("The text node type should not have attributes")
	}
	return result, nil
}

// Like nodes, marks (which are associated with nodes to signify things like
// emphasis or being part of a link) are tagged with type objects, which are
// instantiated once per Schema.
type MarkType struct {
	// The name of the mark type.
	Name string
	Rank int
	// The schema that this mark type instance is part of.
	Schema *Schema
	// The spec on which the type is based.
	Spec     *MarkSpec
	Excluded []*MarkType
	// TODO
}

func NewMarkType(name string, rank int, schema *Schema, spec *MarkSpec) *MarkType {
	return &MarkType{
		Name:   name,
		Rank:   rank,
		Schema: schema,
		Spec:   spec,
		// TODO attrs, excluded, instance
	}
}

// Create a mark of this type. attrs may be null or an object containing only
// some of the mark's attributes. The others, if they have defaults, will be
// added.
func (mt *MarkType) Create(attrs map[string]interface{}) *Mark {
	// TODO if (!attrs && this.instance) return this.instance
	return NewMark(mt, attrs) // TODO computeAttrs
}

func compileMarkType(marks []*MarkSpec, schema *Schema) map[string]*MarkType {
	result := make(map[string]*MarkType)
	for i, m := range marks {
		result[m.Key] = NewMarkType(m.Key, i, schema, m)
	}
	return result
}

// Queries whether a given mark type is excluded by this one.
func (mt *MarkType) Excludes(other *MarkType) bool {
	if len(mt.Excluded) == 0 {
		return false
	}
	for _, ex := range mt.Excluded {
		if other == ex {
			return true
		}
	}
	return false
}

// TODO add other methods to MarkType

// An object describing a schema, as passed to the Schema constructor.
type SchemaSpec struct {
	// The node types in this schema. Maps names to NodeSpec objects that
	// describe the node type associated with that name. Their order is
	// significant—it determines which parse rules take precedence by default,
	// and which nodes come first in a given group.
	Nodes []*NodeSpec

	// The mark types that exist in this schema. The order in which they are
	// provided determines the order in which mark sets are sorted and in which
	// parse rules are tried.
	Marks []*MarkSpec

	// The name of the default top-level node for the schema. Defaults to "doc".
	TopNode string
}

type NodeSpec struct {
	// In JavaScript, the NodeSpec are kept in an OrderedMap. In Go, the map
	// doesn't preserve the order of the keys. Instead, an array is used, and
	// the key is kept here.
	Key string

	// The content expression for this node, as described in the schema guide.
	// When not given, the node does not allow any content.
	Content string

	// The marks that are allowed inside of this node. May be a space-separated
	// string referring to mark names or groups, "_" to explicitly allow all
	// marks, or "" to disallow marks. When not given, nodes with inline
	// content default to allowing all marks, other nodes default to not
	// allowing marks.
	Marks *string

	// The group or space-separated groups to which this node belongs, which
	// can be referred to in the content expressions for the schema.
	Group string

	// Should be set to true for inline nodes. (Implied for text nodes.)
	Inline bool

	// The attributes that nodes of this type get.
	Attrs map[string]*AttributeSpec

	// TODO there are more fields, but are they useful on the server?
}

type MarkSpec struct {
	// In JavaScript, the MarkSpec are kept in an OrderedMap. In Go, the map
	// doesn't preserve the order of the keys. Instead, an array is used, and
	// the key is kept here.
	Key string

	// The attributes that marks of this type get.
	Attrs map[string]*AttributeSpec

	// Whether this mark should be active when the cursor is positioned
	// at its end (or at its start when that is also the start of the
	// parent node). Defaults to true.
	Inclusive bool

	// Determines which other marks this mark can coexist with. Should be a
	// space-separated strings naming other marks or groups of marks. When a
	// mark is added to a set, all marks that it excludes are removed in the
	// process. If the set contains any mark that excludes the new mark but is
	// not, itself, excluded by the new mark, the mark can not be added an the
	// set. You can use the value `"_"` to indicate that the mark excludes all
	// marks in the schema.
	//
	// Defaults to only being exclusive with marks of the same type. You can
	// set it to an empty string (or any string not containing the mark's own
	// name) to allow multiple marks of a given type to coexist (as long as
	// they have different attributes).
	Excludes *string

	// The group or space-separated groups to which this mark belongs.
	Group string

	// TODO there are more fields, but are they useful on the server?
}

// Used to define attributes on nodes or marks.
type AttributeSpec struct {
	// The default value for this attribute, to use when no explicit value is
	// provided. Attributes that have no default must be provided whenever a
	// node or mark of a type that has them is created.
	Default interface{}
}

// A document schema. Holds node and mark type objects for the nodes and marks
// that may occur in conforming documents, and provides functionality for
// creating and deserializing such documents.
type Schema struct {
	// The spec on which the schema is based.
	Spec *SchemaSpec

	// An object mapping the schema's node names to node type objects.
	Nodes map[string]*NodeType

	// A map from mark names to mark type objects.
	Marks map[string]*MarkType
}

// Construct a schema from a schema specification.
// TODO should it take a SchemaSpec or a map[string]interface{} for spec parameter?
func NewSchema(spec *SchemaSpec) (*Schema, error) {
	schema := Schema{
		Spec: spec,
	}
	nodes, err := compileNodeType(spec.Nodes, &schema)
	if err != nil {
		return nil, err
	}
	schema.Nodes = nodes
	schema.Marks = compileMarkType(spec.Marks, &schema)
	// TODO
	for _, typ := range schema.Marks {
		excl := typ.Spec.Excludes
		if excl == nil {
			typ.Excluded = []*MarkType{typ}
		} else if *excl == "" {
			typ.Excluded = []*MarkType{}
		} else {
			gathered, err := gatherMarks(&schema, strings.Fields(*excl))
			if err != nil {
				return nil, err
			}
			typ.Excluded = gathered
		}
	}
	if spec.TopNode == "" {
		spec.TopNode = "doc"
	}
	return &schema, nil
}

func gatherMarks(schema *Schema, marks []string) ([]*MarkType, error) {
	var found []*MarkType
	for _, name := range marks {
		mark, ok := schema.Marks[name]
		if ok {
			found = append(found, mark)
		} else {
			for _, mark = range schema.Marks {
				if name == "_" || hasGroup(mark.Spec.Group, name) {
					found = append(found, mark)
					ok = true
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("Unknown mark type: %s", name)
		}
	}
	return found, nil
}

func hasGroup(groups, group string) bool {
	for _, g := range strings.Fields(groups) {
		if g == group {
			return true
		}
	}
	return false
}
