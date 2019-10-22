package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	strong = &Mark{Type: &MarkType{Name: "strong"}}
	em     = &Mark{Type: &MarkType{Name: "em"}}
	code   = &Mark{Type: &MarkType{Name: "code"}}
)

func TestMarkSameSet(t *testing.T) {
	// returns true for two empty sets
	assert.True(t, SameMarkSet([]*Mark{}, []*Mark{}))

	// returns true for simple identical sets
	assert.True(t, SameMarkSet([]*Mark{em, strong}, []*Mark{em, strong}))

	// returns false for different sets
	assert.False(t, SameMarkSet([]*Mark{em, strong}, []*Mark{em, code}))

	// returns false when set size differs
	assert.False(t, SameMarkSet([]*Mark{em, strong}, []*Mark{em, strong, code}))

	// TODO recognizes links in set
}
