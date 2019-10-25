package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeToString(t *testing.T) {
	// nests
	assert.Equal(t,
		doc(ul(li(p("hey"), p()), li(p("foo")))).String(),
		`doc(bullet_list(list_item(paragraph("hey"), paragraph), list_item(paragraph("foo"))))`,
	)

	// shows inline children
	assert.Equal(t,
		doc(p("foo", img, br, "bar")).String(),
		`doc(paragraph("foo", image, hard_break, "bar"))`,
	)

	// shows marks
	assert.Equal(t,
		doc(p("foo", em("bar", strong("quux")), code("baz"))).String(),
		`doc(paragraph("foo", em("bar"), em(strong("quux")), code("baz")))`,
	)
}
