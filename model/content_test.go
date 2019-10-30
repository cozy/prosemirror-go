package model_test

import (
	"strings"
	"testing"

	. "github.com/cozy/prosemirror-go/model"
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
	for _, t := range strings.Fields(types) {
		ts = append(ts, schema.Nodes[t])
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

// TODO fillBefore
