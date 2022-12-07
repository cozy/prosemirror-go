// Package model implements ProseMirror's document model, along with the
// mechanisms needed to support schemas.
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// For node types where all attrs have a default value (or which don't have any
// attributes), build up a single reusable default attribute object, and use it
// for all nodes that don't specify specific attributes.
func defaultAttrs(attrs map[string]*Attribute) map[string]interface{} {
	defaults := map[string]interface{}{}
	for name, attr := range attrs {
		if !attr.HasDefault {
			return nil
		}
		defaults[name] = attr.Default
	}
	return defaults
}

func computeAttrs(attrs map[string]*Attribute, value ...map[string]interface{}) map[string]interface{} {
	var v map[string]interface{}
	if len(value) > 0 {
		v = value[0]
	}
	built := map[string]interface{}{}
	for name, attr := range attrs {
		given, ok := v[name]
		if !ok {
			if !attr.HasDefault {
				panic(fmt.Errorf("No value supplied for attribute %s", name))
			}
			given = attr.Default
		}
		built[name] = given
	}
	return built
}

func initAttrs(attrs map[string]*AttributeSpec) map[string]*Attribute {
	result := map[string]*Attribute{}
	for name, attr := range attrs {
		result[name] = NewAttribute(attr)
	}
	return result
}

// NodeType are objects allocated once per Schema and used to tag Node
// instances. They contain information about the node type, such as its name
// and what kind of node it represents.
type NodeType struct {
	// The name the node type has in this schema.
	Name string
	// A link back to the `Schema` the node type belongs to.
	Schema *Schema
	// The spec that this type is based on
	Spec         *NodeSpec
	Groups       []string
	Attrs        map[string]*Attribute
	DefaultAttrs map[string]interface{}
	// The starting match of the node type's content expression.
	ContentMatch *ContentMatch
	// The set of marks allowed in this node. `null` means all marks are
	// allowed.
	MarkSet *[]*MarkType
	// True if this node type has inline content.
	InlineContent bool
}

// NewNodeType is the constructor for NodeType.
func NewNodeType(name string, schema *Schema, spec *NodeSpec) *NodeType {
	attrs := initAttrs(spec.Attrs)
	return &NodeType{
		Name:          name,
		Schema:        schema,
		Spec:          spec,
		Groups:        strings.Split(spec.Group, " "),
		Attrs:         attrs,
		DefaultAttrs:  defaultAttrs(attrs),
		ContentMatch:  nil,
		InlineContent: false,
	}
}

// IsText returns true if this is the text node type.
func (nt *NodeType) IsText() bool {
	return nt.Name == "text"
}

// IsBlock returns true if this is a block type.
func (nt *NodeType) IsBlock() bool {
	return !nt.Spec.Inline && nt.Name != "text"
}

// IsInline returns true if this is an inline type.
func (nt *NodeType) IsInline() bool {
	return !nt.IsBlock()
}

// IsLeaf returns true for node types that allow no content.
func (nt *NodeType) IsLeaf() bool {
	return nt.ContentMatch == EmptyContentMatch
}

// IsAtom returns true when this node is an atom, i.e. when it does not have
// directly editable content.
func (nt *NodeType) IsAtom() bool {
	return nt.IsLeaf() || nt.Spec.Atom
}

// HasRequiredAttrs tells you whether this node type has any required
// attributes.
func (nt *NodeType) HasRequiredAttrs() bool {
	for _, attr := range nt.Attrs {
		if attr.isRequired() {
			return true
		}
	}
	return false
}

func (nt *NodeType) compatibleContent(other *NodeType) bool {
	return nt == other || nt.ContentMatch.compatible(other.ContentMatch)
}

func (nt *NodeType) computeAttrs(attrs map[string]interface{}) map[string]interface{} {
	if len(attrs) == 0 && len(nt.DefaultAttrs) > 0 {
		return nt.DefaultAttrs
	}
	return computeAttrs(nt.Attrs, attrs)
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

// CreateChecked is like create, but check the given content against the node
// type's content restrictions, and throw an error if it doesn't match.
//
// :: (?Object, ?union<Fragment, Node, [Node]>, ?[Mark]) → Node
func (nt *NodeType) CreateChecked(args ...interface{}) (*Node, error) {
	var attrs map[string]interface{}
	if len(args) > 0 && args[0] != nil {
		arg, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid type for attrs: %v (%T)", args[0], args[0])
		}
		attrs = arg
	}
	var content interface{}
	if len(args) > 1 {
		content = args[1]
	}
	var marks []*Mark
	if len(args) > 2 && args[2] != nil {
		arg, ok := args[2].([]*Mark)
		if !ok {
			return nil, fmt.Errorf("Invalid type for marks: %v (%T)", args[2], args[2])
		}
		marks = arg
	}

	fragment, err := FragmentFrom(content)
	if err != nil {
		return nil, err
	}
	if !nt.ValidContent(fragment) {
		return nil, fmt.Errorf("Invalid content for node %s", nt.Name)
	}
	return NewNode(nt, nt.computeAttrs(attrs), fragment, MarkSetFrom(marks)), nil
}

// CreateAndFill is like create, but see if it is necessary to add nodes to the
// start or end of the given fragment to make it fit the node. If no fitting
// wrapping can be found, return null. Note that, due to the fact that required
// nodes can always be created, this will always succeed if you pass null or
// Fragment.empty as content.
//
// :: (?Object, ?union<Fragment, Node, [Node]>, ?[Mark]) → ?Node
func (nt *NodeType) CreateAndFill(args ...interface{}) (*Node, error) {
	var attrs map[string]interface{}
	if len(args) > 0 && args[0] != nil {
		arg, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid type for attrs: %v (%T)", args[0], args[0])
		}
		attrs = arg
	}
	var content interface{}
	if len(args) > 1 {
		content = args[1]
	}
	var marks []*Mark
	if len(args) > 2 && args[2] != nil {
		arg, ok := args[2].([]*Mark)
		if !ok {
			return nil, fmt.Errorf("Invalid type for marks: %v (%T)", args[2], args[2])
		}
		marks = arg
	}

	attrs = nt.computeAttrs(attrs)
	fragment, err := FragmentFrom(content)
	if err != nil {
		return nil, err
	}
	if fragment.Size > 0 {
		before := nt.ContentMatch.FillBefore(fragment)
		if before == nil {
			return nil, nil
		}
		fragment = before.Append(fragment)
	}
	after := nt.ContentMatch.MatchFragment(fragment).FillBefore(EmptyFragment, true)
	if after == nil {
		return nil, nil
	}
	return NewNode(nt, attrs, fragment.Append(after), MarkSetFrom(marks)), nil
}

// ValidContent returns true if the given fragment is valid content for this
// node type with the given attributes.
func (nt *NodeType) ValidContent(content *Fragment) bool {
	result := nt.ContentMatch.MatchFragment(content)
	if result == nil || !result.ValidEnd {
		return false
	}
	for _, child := range content.Content {
		if !nt.AllowsMarks(child.Marks) {
			return false
		}
	}
	return true
}

// AllowsMarkType checks whether the given mark type is allowed in this node.
func (nt *NodeType) AllowsMarkType(markType *MarkType) bool {
	if nt.MarkSet == nil {
		return true
	}
	for _, mt := range *nt.MarkSet {
		if mt == markType {
			return true
		}
	}
	return false
}

// AllowsMarks tests whether the given set of marks are allowed in this node.
func (nt *NodeType) AllowsMarks(marks []*Mark) bool {
	if nt.MarkSet == nil {
		return true
	}
	for _, mark := range marks {
		if !nt.AllowsMarkType(mark.Type) {
			return false
		}
	}
	return true
}

func findNoteType(types []*NodeType, key string) (*NodeType, bool) {
	for _, t := range types {
		if t.Name == key {
			return t, true
		}
	}
	return nil, false
}

func compileNodeType(nodes []*NodeSpec, schema *Schema) ([]*NodeType, error) {
	var result []*NodeType
	for _, n := range nodes {
		nt := NewNodeType(n.Key, schema, n)
		result = append(result, nt)
	}
	topType := schema.Spec.TopNode
	if _, ok := findNoteType(result, topType); !ok {
		return nil, fmt.Errorf("The schema is missing its top node type (%s)", topType)
	}
	txt, ok := findNoteType(result, "text")
	if !ok {
		return nil, errors.New("Every schema needs a 'text' type")
	}
	if len(txt.Attrs) > 0 {
		return nil, errors.New("The text node type should not have attributes")
	}
	return result, nil
}

// Attribute descriptors
type Attribute struct {
	HasDefault bool
	Default    interface{}
}

func (a *Attribute) isRequired() bool {
	return !a.HasDefault
}

// NewAttribute is the constructor for Attribute.
func NewAttribute(options *AttributeSpec) *Attribute {
	if options == nil {
		return &Attribute{HasDefault: false, Default: nil}
	}
	return &Attribute{HasDefault: true, Default: options.Default}
}

// MarkType is the type object for marks. Like nodes, marks (which are
// associated with nodes to signify things like emphasis or being part of a
// link) are tagged with type objects, which are instantiated once per Schema.
type MarkType struct {
	// The name of the mark type.
	Name string
	Rank int
	// The schema that this mark type instance is part of.
	Schema *Schema
	// The spec on which the type is based.
	Spec     *MarkSpec
	Excluded []*MarkType
	Attrs    map[string]*Attribute
	Instance *Mark
}

// NewMarkType is the constructor for MarkType.
func NewMarkType(name string, rank int, schema *Schema, spec *MarkSpec) *MarkType {
	mt := &MarkType{
		Name:   name,
		Rank:   rank,
		Schema: schema,
		Spec:   spec,
		Attrs:  initAttrs(spec.Attrs),
	}
	defaults := defaultAttrs(mt.Attrs)
	if len(defaults) > 0 {
		mt.Instance = NewMark(mt, defaults)
	}
	return mt
}

// Create a mark of this type. attrs may be null or an object containing only
// some of the mark's attributes. The others, if they have defaults, will be
// added.
func (mt *MarkType) Create(attrs map[string]interface{}) *Mark {
	if len(mt.Attrs) == 0 && mt.Instance != nil {
		return mt.Instance
	}
	return NewMark(mt, computeAttrs(mt.Attrs, attrs))
}

func compileMarkType(marks []*MarkSpec, schema *Schema) []*MarkType {
	var result []*MarkType
	for i, m := range marks {
		mt := NewMarkType(m.Key, i, schema, m)
		result = append(result, mt)
	}
	return result
}

// IsInSet tests whether there is a mark of this type in the given set.
func (mt *MarkType) IsInSet(set []*Mark) *Mark {
	for _, mark := range set {
		if mark.Type == mt {
			return mark
		}
	}
	return nil
}

// Excludes queries whether a given mark type is excluded by this one.
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

func findMarkType(types []*MarkType, key string) (*MarkType, bool) {
	for _, t := range types {
		if t.Name == key {
			return t, true
		}
	}
	return nil, false
}

// SchemaSpec is an object describing a schema, as passed to the Schema
// constructor.
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

// SchemaSpecFromJSON returns a SchemaSpec from a JSON representation.
func SchemaSpecFromJSON(raw map[string]interface{}) SchemaSpec {
	var spec SchemaSpec

	if nodes, ok := raw["nodes"].([]interface{}); ok {
		for _, node := range nodes {
			if tuple, ok := node.([]interface{}); ok && len(tuple) == 2 {
				n := &NodeSpec{}
				if key, ok := tuple[0].(string); ok {
					n.Key = key
				}
				data, ok := tuple[1].(map[string]interface{})
				if !ok {
					continue
				}
				if content, ok := data["content"].(string); ok {
					n.Content = content
				}
				if marks, ok := data["marks"].(string); ok {
					n.Marks = &marks
				}
				if group, ok := data["group"].(string); ok {
					n.Group = group
				}
				if inline, ok := data["inline"].(bool); ok {
					n.Inline = inline
				}
				if attrs, ok := data["attrs"].(map[string]interface{}); ok {
					n.Attrs = make(map[string]*AttributeSpec)
					for k, v := range attrs {
						attr, _ := v.(map[string]interface{})
						n.Attrs[k] = &AttributeSpec{Default: attr["default"]}
					}
				}
				spec.Nodes = append(spec.Nodes, n)
			}
		}
	}

	if marks, ok := raw["marks"].([]interface{}); ok {
		for _, mark := range marks {
			if tuple, ok := mark.([]interface{}); ok && len(tuple) == 2 {
				m := &MarkSpec{}
				if key, ok := tuple[0].(string); ok {
					m.Key = key
				}
				data, ok := tuple[1].(map[string]interface{})
				if !ok {
					continue
				}
				if attrs, ok := data["attrs"].(map[string]interface{}); ok {
					m.Attrs = make(map[string]*AttributeSpec)
					for k, v := range attrs {
						attr, _ := v.(map[string]interface{})
						m.Attrs[k] = &AttributeSpec{Default: attr["default"]}
					}
				}
				if incl, ok := data["inclusive"].(bool); ok {
					m.Inclusive = &incl
				}
				if excl, ok := data["excludes"].(string); ok {
					m.Excludes = &excl
				}
				if group, ok := data["group"].(string); ok {
					m.Group = group
				}
				spec.Marks = append(spec.Marks, m)
			}
		}
	}

	spec.TopNode, _ = raw["topNode"].(string)
	return spec
}

// MarshalJSON creates a JSON representation of the SchemaSpec.
func (s SchemaSpec) MarshalJSON() ([]byte, error) {
	if len(s.Nodes) == 0 && len(s.Marks) == 0 && len(s.TopNode) == 0 {
		return []byte(`{}`), nil
	}

	buf := make([]byte, 0, 4096)
	if len(s.Nodes) > 0 {
		buf = append(buf, []byte(`,"nodes":[`)...)
		for _, node := range s.Nodes {
			val, err := json.Marshal(node)
			if err != nil {
				return nil, err
			}
			buf = append(buf, []byte(`["`+node.Key+`",`)...)
			buf = append(buf, val...)
			buf = append(buf, ']', ',')
		}
		buf[len(buf)-1] = ']'
	}

	if len(s.Marks) > 0 {
		buf = append(buf, []byte(`,"marks":[`)...)
		for _, mark := range s.Marks {
			val, err := json.Marshal(mark)
			if err != nil {
				return nil, err
			}
			buf = append(buf, []byte(`["`+mark.Key+`",`)...)
			buf = append(buf, val...)
			buf = append(buf, ']', ',')
		}
		buf[len(buf)-1] = ']'
	}

	if len(s.TopNode) > 0 {
		buf = append(buf, []byte(`,"topNode":"`+s.TopNode+`"`)...)
	}
	buf[0] = '{'
	buf = append(buf, '}')
	return buf, nil
}

// UnmarshalJSON parses a JSON representation of a SchemaSpec.
func (s *SchemaSpec) UnmarshalJSON(buf []byte) error {
	var raw struct {
		Nodes   [][2]json.RawMessage `json:"nodes"`
		Marks   [][2]json.RawMessage `json:"marks"`
		TopNode string               `json:"topNode"`
	}
	if err := json.Unmarshal(buf, &raw); err != nil {
		return err
	}

	s.Nodes = s.Nodes[:0]
	for _, n := range raw.Nodes {
		var node NodeSpec
		if err := json.Unmarshal(n[1], &node); err != nil {
			return err
		}
		if len(n[0]) < 2 {
			return errors.New("Invalid node key")
		}
		key := []byte(n[0])
		node.Key = string(key[1 : len(key)-1]) // Remove the "" around the key
		s.Nodes = append(s.Nodes, &node)
	}

	s.Marks = s.Marks[:0]
	for _, m := range raw.Marks {
		var mark MarkSpec
		if err := json.Unmarshal(m[1], &mark); err != nil {
			return err
		}
		if len(m[0]) < 2 {
			return errors.New("Invalid mark key")
		}
		key := []byte(m[0])
		mark.Key = string(key[1 : len(key)-1]) // Remove the "" around the key
		s.Marks = append(s.Marks, &mark)
	}

	s.TopNode = raw.TopNode
	return nil
}

// NodeSpec is an object describing a node type.
type NodeSpec struct {
	// In JavaScript, the NodeSpec are kept in an OrderedMap. In Go, the map
	// doesn't preserve the order of the keys. Instead, an array is used, and
	// the key is kept here.
	Key string `json:"-"`

	// The content expression for this node, as described in the schema guide.
	// When not given, the node does not allow any content.
	Content string `json:"content,omitempty"`

	// The marks that are allowed inside of this node. May be a space-separated
	// string referring to mark names or groups, "_" to explicitly allow all
	// marks, or "" to disallow marks. When not given, nodes with inline
	// content default to allowing all marks, other nodes default to not
	// allowing marks.
	Marks *string `json:"marks,omitempty"`

	// The group or space-separated groups to which this node belongs, which
	// can be referred to in the content expressions for the schema.
	Group string `json:"group,omitempty"`

	// Should be set to true for inline nodes. (Implied for text nodes.)
	Inline bool `json:"inline,omitempty"`

	// Can be set to true to indicate that, though this isn't a [leaf
	// node](#model.NodeType.isLeaf), it doesn't have directly editable
	// content and should be treated as a single unit in the view.
	Atom bool `json:"atom,omitempty"`

	// The attributes that nodes of this type get.
	Attrs map[string]*AttributeSpec `json:"attrs,omitempty"`

	// Defines the default way a node of this type should be serialized to a
	// string representation for debugging (e.g. in error messages).
	ToDebugString func(*Node) string `json:"-"`
}

// MarkSpec is an object describing a mark type.
type MarkSpec struct {
	// In JavaScript, the MarkSpec are kept in an OrderedMap. In Go, the map
	// doesn't preserve the order of the keys. Instead, an array is used, and
	// the key is kept here.
	Key string `json:"-"`

	// The attributes that marks of this type get.
	Attrs map[string]*AttributeSpec `json:"attrs,omitempty"`

	// Whether this mark should be active when the cursor is positioned
	// at its end (or at its start when that is also the start of the
	// parent node). Defaults to true.
	Inclusive *bool `json:"inclusive,omitempty"`

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
	Excludes *string `json:"excludes,omitempty"`

	// The group or space-separated groups to which this mark belongs.
	Group string `json:"group,omitempty"`
}

// AttributeSpec is used to define attributes on nodes or marks.
type AttributeSpec struct {
	// The default value for this attribute, to use when no explicit value is
	// provided. Attributes that have no default must be provided whenever a
	// node or mark of a type that has them is created.
	Default interface{} `json:"default,omitempty"`
}

// Schema is a a document schema: it holds node and mark type objects for the
// nodes and marks that may occur in conforming documents, and provides
// functionality for creating and deserializing such documents.
type Schema struct {
	// The spec on which the schema is based.
	Spec *SchemaSpec

	// An object mapping the schema's node names to node type objects.
	Nodes []*NodeType

	// A map from mark names to mark type objects.
	Marks []*MarkType
}

// NewSchema constructs a schema from a schema specification.
func NewSchema(spec *SchemaSpec) (*Schema, error) {
	schema := Schema{
		Spec: spec,
	}
	if spec.TopNode == "" {
		spec.TopNode = "doc"
	}
	nodes, err := compileNodeType(spec.Nodes, &schema)
	if err != nil {
		return nil, err
	}
	schema.Nodes = nodes
	schema.Marks = compileMarkType(spec.Marks, &schema)

	contentExprCache := map[string]*ContentMatch{}
	for _, typ := range schema.Nodes {
		if _, ok := findMarkType(schema.Marks, typ.Name); ok {
			return nil, fmt.Errorf("%s can not be both a node and a mark", typ.Name)
		}
		contentExpr := typ.Spec.Content
		markExpr := typ.Spec.Marks
		cm, ok := contentExprCache[contentExpr]
		if !ok {
			cm, err = ParseContentMatch(contentExpr, schema.Nodes)
			if err != nil {
				return nil, err
			}
			contentExprCache[contentExpr] = cm
		}
		typ.ContentMatch = cm
		typ.InlineContent = typ.ContentMatch.inlineContent()
		if markExpr == nil {
			if !typ.InlineContent {
				var set []*MarkType
				typ.MarkSet = &set
			}
		} else if *markExpr == "" {
			var set []*MarkType
			typ.MarkSet = &set
		} else if *markExpr == "_" {
			typ.MarkSet = nil
		} else {
			set, err := gatherMarks(&schema, strings.Split(*markExpr, " "))
			if err != nil {
				return nil, err
			}
			typ.MarkSet = &set
		}
	}

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
	return &schema, nil
}

// Node creates a node in this schema. The type may be a string or a NodeType
// instance. Attributes will be extended with defaults, content may be a
// Fragment, null, a Node, or an array of nodes.
//
// :: (union<string, NodeType>, ?Object, ?union<Fragment, Node, [Node]>, ?[Mark]) → Node
func (s *Schema) Node(typ interface{}, args ...interface{}) (*Node, error) {
	var t *NodeType
	switch typ := typ.(type) {
	case *NodeType:
		t = typ
		if t.Schema != s {
			return nil, fmt.Errorf("Node type from different schema used (%s)", t.Name)
		}
	case string:
		var err error
		t, err = s.NodeType(typ)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid node type: %v (%T)", typ, typ)
	}
	var attrs map[string]interface{}
	if len(args) > 0 && args[0] != nil {
		arg, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid type for attrs: %v (%T)", args[0], args[0])
		}
		attrs = arg
	}
	var content interface{}
	if len(args) > 1 {
		content = args[1]
	}
	var marks []*Mark
	if len(args) > 2 && args[2] != nil {
		arg, ok := args[2].([]*Mark)
		if !ok {
			return nil, fmt.Errorf("Invalid type for marks: %v (%T)", args[2], args[2])
		}
		marks = arg
	}
	return t.CreateChecked(attrs, content, marks)
}

// Text creates a text node in the schema. Empty text nodes are not allowed.
func (s *Schema) Text(text string, marks ...[]*Mark) *Node {
	typ, ok := findNoteType(s.Nodes, "text")
	if !ok {
		panic(errors.New("No text node type"))
	}
	set := NoMarks
	if len(marks) > 0 {
		set = MarkSetFrom(marks[0])
	}
	return NewTextNode(typ, typ.DefaultAttrs, text, set)
}

// Mark creates a mark with the given type and attributes.
func (s *Schema) Mark(typ interface{}, args ...map[string]interface{}) *Mark {
	var t *MarkType
	switch typ := typ.(type) {
	case *MarkType:
		t = typ
	case string:
		t, _ = findMarkType(s.Marks, typ)
	}
	var attrs map[string]interface{}
	if len(args) > 0 {
		attrs = args[0]
	}
	return t.Create(attrs)
}

// NodeFromJSON deserializes a node from its JSON representation.
func (s *Schema) NodeFromJSON(raw []byte) (*Node, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return NodeFromJSON(s, obj)
}

// MarkFromJSON deserializes a mark from its JSON representation.
func (s *Schema) MarkFromJSON(raw []byte) (*Mark, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return MarkFromJSON(s, obj)
}

// NodeType returns the NodeType with the given name in this schema.
func (s *Schema) NodeType(name string) (*NodeType, error) {
	if found, ok := findNoteType(s.Nodes, name); ok {
		return found, nil
	}
	return nil, fmt.Errorf("Unknown node type: %s", name)
}

// MarkType returns the MarkType with the given name in this schema.
func (s *Schema) MarkType(name string) (*MarkType, error) {
	if found, ok := findMarkType(s.Marks, name); ok {
		return found, nil
	}
	return nil, fmt.Errorf("Unknown mark type: %s", name)
}

func gatherMarks(schema *Schema, marks []string) ([]*MarkType, error) {
	var found []*MarkType
	for _, name := range marks {
		mark, ok := findMarkType(schema.Marks, name)
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
