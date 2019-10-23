package model

// Like nodes, marks (which are associated with nodes to signify things like
// emphasis or being part of a link) are tagged with type objects, which are
// instantiated once per Schema.
type MarkType struct {
	Name string
	Rank int
	// TODO
}

// Queries whether a given mark type is excluded by this one.
func (mt *MarkType) Excludes(other *MarkType) bool {
	return false // TODO
}

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
	Excludes string

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

// TODO
