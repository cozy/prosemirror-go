package model_test

import (
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/stretchr/testify/assert"
)

func TestMarkSameSet(t *testing.T) {
	// returns true for two empty sets
	assert.True(t, SameMarkSet([]*Mark{}, []*Mark{}))

	// returns true for simple identical sets
	assert.True(t, SameMarkSet([]*Mark{em2, strong2}, []*Mark{em2, strong2}))

	// returns false for different sets
	assert.False(t, SameMarkSet([]*Mark{em2, strong2}, []*Mark{em2, code}))

	// returns false when set size differs
	assert.False(t, SameMarkSet([]*Mark{em2, strong2}, []*Mark{em2, strong2, code}))

	// recognizes identical links in set
	assert.True(t, SameMarkSet(
		[]*Mark{link("http://foo"), code},
		[]*Mark{link("http://foo"), code}))

	// recognizes different links in set
	assert.False(t, SameMarkSet(
		[]*Mark{link("http://foo"), code},
		[]*Mark{link("http://bar"), code}))
}

func TestMarkEq(t *testing.T) {
	// considers identical links to be the same
	assert.True(t, link("http://foo").Eq(link("http://foo")))

	// considers different links to differ
	assert.False(t, link("http://foo").Eq(link("http://bar")))

	// considers links with different titles to differ
	assert.False(t, link("http://foo").Eq(link("http://foo", "B")))
}

func TestMarkAddToSet(t *testing.T) {
	customSchema, err := NewSchema(&SchemaSpec{
		Nodes: []*NodeSpec{
			{Key: "doc", Content: "paragraph+"},
			{Key: "paragraph", Content: "text*"},
			{Key: "text"},
		},
		Marks: []*MarkSpec{
			{Key: "remark", Attrs: idAttrs, Excludes: &empty, Inclusive: &falsy},
			{Key: "user", Attrs: idAttrs, Excludes: &underscore},
			{Key: "strong2", Excludes: &emGroup},
			{Key: "em", Group: emGroup},
		},
	})
	assert.NoError(t, err)
	custom := customSchema.Marks

	remark1 := custom["remark"].Create(map[string]interface{}{"id": 1})
	remark2 := custom["remark"].Create(map[string]interface{}{"id": 2})
	user1 := custom["user"].Create(map[string]interface{}{"id": 1})
	user2 := custom["user"].Create(map[string]interface{}{"id": 2})
	customEm := custom["em"].Create(nil)
	customStrong := custom["strong2"].Create(nil)

	// can add to the empty set
	assert.True(t, SameMarkSet(
		em2.AddToSet([]*Mark{}),
		[]*Mark{em2},
	))

	// is a no-op when the added thing is in set
	assert.True(t, SameMarkSet(
		em2.AddToSet([]*Mark{em2}),
		[]*Mark{em2},
	))

	// adds marks with lower rank before others
	assert.True(t, SameMarkSet(
		em2.AddToSet([]*Mark{strong2}),
		[]*Mark{em2, strong2},
	))

	// adds marks with higher rank after others
	assert.True(t, SameMarkSet(
		strong2.AddToSet([]*Mark{em2}),
		[]*Mark{em2, strong2},
	))

	// replaces different marks with new attributes
	assert.True(t, SameMarkSet(
		link("http://bar").AddToSet([]*Mark{link("http://foo"), em2}),
		[]*Mark{link("http://bar"), em2},
	))

	// does nothing when adding an existing link
	assert.True(t, SameMarkSet(
		link("http://foo").AddToSet([]*Mark{em2, link("http://foo")}),
		[]*Mark{em2, link("http://foo")},
	))

	// puts code marks at the end
	assert.True(t, SameMarkSet(
		code.AddToSet([]*Mark{em2, strong2, link("http://foo")}),
		[]*Mark{em2, strong2, link("http://foo"), code},
	))

	// puts marks with middle rank in the middle
	assert.True(t, SameMarkSet(
		strong2.AddToSet([]*Mark{em2, code}),
		[]*Mark{em2, strong2, code},
	))

	// allows nonexclusive instances of marks with the same type
	assert.True(t, SameMarkSet(
		remark2.AddToSet([]*Mark{remark1}),
		[]*Mark{remark1, remark2},
	))

	// doesn't duplicate identical instances of nonexclusive marks
	assert.True(t, SameMarkSet(
		remark1.AddToSet([]*Mark{remark1}),
		[]*Mark{remark1},
	))

	// clears all others when adding a globally-excluding mark
	assert.True(t, SameMarkSet(
		user1.AddToSet([]*Mark{remark1, customEm}),
		[]*Mark{user1},
	))

	// does not allow adding another mark to a globally-excluding mark
	assert.True(t, SameMarkSet(
		customEm.AddToSet([]*Mark{user1}),
		[]*Mark{user1},
	))

	// does overwrite a globally-excluding mark when adding another instance
	assert.True(t, SameMarkSet(
		user2.AddToSet([]*Mark{user1}),
		[]*Mark{user2},
	))

	// doesn't add anything when another mark excludes the added mark
	assert.True(t, SameMarkSet(
		customEm.AddToSet([]*Mark{remark1, customStrong}),
		[]*Mark{remark1, customStrong},
	))

	// remove excluded marks when adding a mark
	assert.True(t, SameMarkSet(
		customStrong.AddToSet([]*Mark{remark1, customEm}),
		[]*Mark{remark1, customStrong},
	))
}

func TestMarkRemoveFromSet(t *testing.T) {
	// is a no-op for the empty set
	assert.True(t, SameMarkSet(em2.RemoveFromSet([]*Mark{}), []*Mark{}))

	// can remove the last mark from a set
	assert.True(t, SameMarkSet(em2.RemoveFromSet([]*Mark{em2}), []*Mark{}))

	// is a no-op when the mark isn't in the set
	assert.True(t, SameMarkSet(strong2.RemoveFromSet([]*Mark{em2}), []*Mark{em2}))

	// can remove a mark with attributes
	assert.True(t, SameMarkSet(
		link("http://foo").RemoveFromSet([]*Mark{link("http://foo")}),
		[]*Mark{},
	))

	// doesn't remove a mark when its attrs differ
	// can remove a mark with attributes
	assert.True(t, SameMarkSet(
		link("http://foo", "title").RemoveFromSet([]*Mark{link("http://foo")}),
		[]*Mark{link("http://foo")},
	))
}

// TODO ResolvedPos.marks
