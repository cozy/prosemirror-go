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

// TODO
