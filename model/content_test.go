package model_test

import (
	"strings"
	"testing"

	. "github.com/shodgson/prosemirror-go/model"
	"github.com/shodgson/prosemirror-go/test/builder"
	"github.com/stretchr/testify/assert"
)

func get(t *testing.T, expr string) *ContentMatch {
	cm, err := ParseContentMatch(expr, schema.Nodes)
	assert.NoError(t, err)
	return cm
}

func match(t *testing.T, expr, types string) bool {
	m := get(t, expr)
	var ts []*NodeType
	for _, name := range strings.Fields(types) {
		typ, err := schema.NodeType(name)
		assert.NoError(t, err)
		ts = append(ts, typ)
	}
	for i := 0; m != nil && i < len(ts); i++ {
		m = m.MatchType(ts[i])
	}
	return m != nil && m.ValidEnd
}

func valid(t *testing.T, expr, types string) {
	assert.True(t, match(t, expr, types))
}
func invalid(t *testing.T, expr, types string) {
	assert.False(t, match(t, expr, types))
}

func fill(t *testing.T, expr string, before, after builder.NodeWithTag, result interface{}) {
	filled := get(t, expr).MatchFragment(before.Content).FillBefore(after.Content, true)
	if result != nil {
		if assert.NotNil(t, filled) {
			content := result.(builder.NodeWithTag).Content
			assert.True(t, filled.Eq(content), "%s != %s", filled, content)
		}
	} else {
		assert.Nil(t, filled)
	}
}

func fill3(t *testing.T, expr string, before, mid, after builder.NodeWithTag, args ...interface{}) {
	content := get(t, expr)
	a := content.MatchFragment(before.Content).FillBefore(mid.Content)
	var b *Fragment
	if a != nil {
		b = content.MatchFragment(before.Content.Append(a).Append(mid.Content)).FillBefore(after.Content, true)
	}
	if len(args) > 1 && args[0] != nil {
		expected := args[0].(builder.NodeWithTag).Content
		assert.True(t, a.Eq(expected), "%s != %s\n", a.String(), expected.String())
		expected = args[1].(builder.NodeWithTag).Content
		assert.True(t, b.Eq(expected), "%s != %s\n", b.String(), expected.String())
	} else {
		assert.Nil(t, b)
	}
}

func TestContentMatchMatchType(t *testing.T) {
	// accepts empty content for the empty expr
	valid(t, "", "")
	// doesn't accept content in the empty expr
	invalid(t, "", "image")

	// matches nothing to an asterisk
	valid(t, "image*", "")
	// matches one element to an asterisk
	valid(t, "image*", "image")
	// matches multiple elements to an asterisk
	valid(t, "image*", "image image image image")
	// only matches appropriate elements to an asterisk
	invalid(t, "image*", "image text")

	// matches group members to a group
	valid(t, "inline*", "image text")
	// doesn't match non-members to a group
	invalid(t, "inline*", "paragraph")
	// matches an element to a choice expression
	valid(t, "(paragraph | heading)", "paragraph")
	// doesn't match unmentioned elements to a choice expr
	invalid(t, "(paragraph | heading)", "image")

	// matches a simple sequence
	valid(t, "paragraph horizontal_rule paragraph", "paragraph horizontal_rule paragraph")
	// fails when a sequence is too long
	invalid(t, "paragraph horizontal_rule", "paragraph horizontal_rule paragraph")
	// fails when a sequence is too short
	invalid(t, "paragraph horizontal_rule paragraph", "paragraph horizontal_rule")
	// fails when a sequence starts incorrectly
	invalid(t, "paragraph horizontal_rule", "horizontal_rule paragraph horizontal_rule")

	// accepts a sequence asterisk matching zero elements
	valid(t, "heading paragraph*", "heading")
	// accepts a sequence asterisk matching multiple elts
	valid(t, "heading paragraph*", "heading paragraph paragraph")
	// accepts a sequence plus matching one element
	valid(t, "heading paragraph+", "heading paragraph")
	// accepts a sequence plus matching multiple elts
	valid(t, "heading paragraph+", "heading paragraph paragraph")
	// fails when a sequence plus has no elements
	invalid(t, "heading paragraph+", "heading")
	// fails when a sequence plus misses its start
	invalid(t, "heading paragraph+", "paragraph paragraph")

	// accepts an optional element being present
	valid(t, "image?", "image")
	// accepts an optional element being missing
	valid(t, "image?", "")
	// fails when an optional element is present twice
	invalid(t, "image?", "image image")

	// accepts a nested repeat
	valid(t, "(heading paragraph+)+", "heading paragraph heading paragraph paragraph")
	// fails on extra input after a nested repeat
	invalid(t, "(heading paragraph+)+", "heading paragraph heading paragraph paragraph horizontal_rule")

	// accepts a matching count
	valid(t, "hard_break{2}", "hard_break hard_break")
	// rejects a count that comes up short
	invalid(t, "hard_break{2}", "hard_break")
	// rejects a count that has too many elements
	invalid(t, "hard_break{2}", "hard_break hard_break hard_break")
	// accepts a count on the lower bound
	valid(t, "hard_break{2, 4}", "hard_break hard_break")
	// accepts a count on the upper bound
	valid(t, "hard_break{2, 4}", "hard_break hard_break hard_break hard_break")
	// accepts a count between the bounds
	valid(t, "hard_break{2, 4}", "hard_break hard_break hard_break")
	// rejects a sequence with too few elements
	invalid(t, "hard_break{2, 4}", "hard_break")
	// rejects a sequence with too many elements
	invalid(t, "hard_break{2, 4}", "hard_break hard_break hard_break hard_break hard_break")
	// rejects a sequence with a bad element after it
	invalid(t, "hard_break{2, 4} text*", "hard_break hard_break image")
	// accepts a sequence with a matching element after it
	valid(t, "hard_break{2, 4} image?", "hard_break hard_break image")
	// accepts an open range
	valid(t, "hard_break{2,}", "hard_break hard_break")
	// accepts an open range matching many
	valid(t, "hard_break{2,}", "hard_break hard_break hard_break hard_break")
	// rejects an open range with too few elements
	invalid(t, "hard_break{2,}", "hard_break")
}

func TestContentMatchFillBefore(t *testing.T) {
	// returns the empty fragment when things match
	fill(t, "paragraph horizontal_rule paragraph", doc(p(), hr), doc(p()), doc())

	// adds a node when necessary
	fill(t, "paragraph horizontal_rule paragraph", doc(p()), doc(p()), doc(hr))

	// accepts an asterisk across the bound
	fill(t, "hard_break*", p(br), p(br), p())

	// accepts an asterisk only on the left
	fill(t, "hard_break*", p(br), p(), p())

	// accepts an asterisk only on the right
	fill(t, "hard_break*", p(), p(br), p())

	// accepts an asterisk with no elements
	fill(t, "hard_break*", p(), p(), p())

	// accepts a plus across the bound
	fill(t, "hard_break+", p(br), p(br), p())

	// adds an element for a content-less plus
	fill(t, "hard_break+", p(), p(), p(br))

	// fails for a mismatched plus
	fill(t, "hard_break+", p(), p(img), nil)

	// accepts asterisk with content on both sides
	fill(t, "heading* paragraph*", doc(h1()), doc(p()), doc())

	// accepts asterisk with no content after
	fill(t, "heading* paragraph*", doc(h1()), doc(), doc())

	// accepts plus with content on both sides
	fill(t, "heading+ paragraph+", doc(h1()), doc(p()), doc())

	// accepts plus with no content after
	fill(t, "heading+ paragraph+", doc(h1()), doc(), doc(p()))

	// adds elements to match a count
	fill(t, "hard_break{3}", p(br), p(br), p(br))

	// fails when there are too many elements
	fill(t, "hard_break{3}", p(br, br), p(br, br), nil)

	// adds elements for two counted groups
	fill(t, "code_block{2} paragraph{2}", doc(pre()), doc(p()), doc(pre(), p()))

	// doesn't include optional elements
	fill(t, "heading paragraph? horizontal_rule", doc(h1()), doc(), doc(hr))

	// completes a sequence
	fill3(t, "paragraph horizontal_rule paragraph horizontal_rule paragraph",
		doc(p()), doc(p()), doc(p()), doc(hr), doc(hr))

	// accepts plus across two bounds
	fill3(t, "code_block+ paragraph+",
		doc(pre()), doc(pre()), doc(p()), doc(), doc())

	// fills a plus from empty input
	fill3(t, "code_block+ paragraph+",
		doc(), doc(), doc(), doc(), doc(pre(), p()))

	// completes a count
	fill3(t, "code_block{3} paragraph{3}",
		doc(pre()), doc(p()), doc(), doc(pre(), pre()), doc(p(), p()))

	// fails on non-matching elements
	fill3(t, "paragraph*", doc(p()), doc(pre()), doc(p()), nil)

	// completes a plus across two bounds
	fill3(t, "paragraph{4}", doc(p()), doc(p()), doc(p()), doc(), doc(p()))

	// refuses to complete an overflown count across two bounds
	fill3(t, "paragraph{2}", doc(p()), doc(p()), doc(p()), nil)
}
