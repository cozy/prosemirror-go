package model_test

import (
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/cozy/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

var (
	schema = builder.Schema
	// doc    = builder.Doc
	// p      = builder.P
	// em     = builder.Em
	// a      = builder.A

	strong = schema.Mark("strong")
	em2    = schema.Mark("em")
	code   = schema.Mark("code")
	link   = func(href string, title ...string) *Mark {
		attrs := map[string]interface{}{"href": href}
		if len(title) > 0 {
			attrs["title"] = title[0]
		}
		return schema.Mark("link", attrs)
	}
)

func TestMarkSameSet(t *testing.T) {
	// returns true for two empty sets
	assert.True(t, SameMarkSet([]*Mark{}, []*Mark{}))

	// returns true for simple identical sets
	assert.True(t, SameMarkSet([]*Mark{em2, strong}, []*Mark{em2, strong}))

	// returns false for different sets
	assert.False(t, SameMarkSet([]*Mark{em2, strong}, []*Mark{em2, code}))

	// returns false when set size differs
	assert.False(t, SameMarkSet([]*Mark{em2, strong}, []*Mark{em2, strong, code}))

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
		em2.AddToSet([]*Mark{strong}),
		[]*Mark{em2, strong},
	))

	// adds marks with higher rank after others
	assert.True(t, SameMarkSet(
		strong.AddToSet([]*Mark{em2}),
		[]*Mark{em2, strong},
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
		code.AddToSet([]*Mark{em2, strong, link("http://foo")}),
		[]*Mark{em2, strong, link("http://foo"), code},
	))

	// puts marks with middle rank in the middle
	assert.True(t, SameMarkSet(
		strong.AddToSet([]*Mark{em2, code}),
		[]*Mark{em2, strong, code},
	))

	// TODO custom elements

	// allows nonexclusive instances of marks with the same type
	// ist(remark2.addToSet([remark1]), [remark1, remark2], Mark.sameSet))

	// doesn't duplicate identical instances of nonexclusive marks
	// ist(remark1.addToSet([remark1]), [remark1], Mark.sameSet))

	// clears all others when adding a globally-excluding mark
	// ist(user1.addToSet([remark1, customEm]), [user1], Mark.sameSet))

	// does not allow adding another mark to a globally-excluding mark
	// ist(customEm.addToSet([user1]), [user1], Mark.sameSet))

	// does overwrite a globally-excluding mark when adding another instance
	// ist(user2.addToSet([user1]), [user2], Mark.sameSet))

	// doesn't add anything when another mark excludes the added mark
	// ist(customEm.addToSet([remark1, customStrong]), [remark1, customStrong], Mark.sameSet))

	// remove excluded marks when adding a mark
	// ist(customStrong.addToSet([remark1, customEm]), [remark1, customStrong], Mark.sameSet))
}
