package model_test

import (
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/stretchr/testify/assert"
)

type res struct {
	node  *Node
	start int
	end   int
}

func TestNodeResolve(t *testing.T) {
	testDoc := doc(p("ab"), blockquote(p(em("cd"), "ef")))
	rdoc := res{node: testDoc.Node, start: 0, end: 12}
	child, err := testDoc.Child(0)
	assert.NoError(t, err)
	p1 := res{node: child, start: 1, end: 3}
	child, err = testDoc.Child(1)
	assert.NoError(t, err)
	blk := res{node: child, start: 5, end: 11}
	child, err = blk.node.Child(0)
	assert.NoError(t, err)
	p2 := res{node: child, start: 6, end: 10}

	// It should reflect the document structure
	expected := [][]interface{}{
		{rdoc, 0, nil, p1.node},
		{rdoc, p1, 0, nil, "ab"},
		{rdoc, p1, 1, "a", "b"},
		{rdoc, p1, 2, "ab", nil},
		{rdoc, 4, p1.node, blk.node},
		{rdoc, blk, 0, nil, p2.node},
		{rdoc, blk, p2, 0, nil, "cd"},
		{rdoc, blk, p2, 1, "c", "d"},
		{rdoc, blk, p2, 2, "cd", "ef"},
		{rdoc, blk, p2, 3, "e", "f"},
		{rdoc, blk, p2, 4, "ef", nil},
		{rdoc, blk, 6, p2.node, nil},
		{rdoc, 12, blk.node, nil},
	}

	for pos := 0; pos <= testDoc.Content.Size; pos++ {
		dpos, err := testDoc.Resolve(pos)
		assert.NoError(t, err)
		exp := expected[pos]
		assert.Equal(t, dpos.Depth, len(exp)-4)
		for i := 0; i < len(exp)-3; i++ {
			ex := exp[i].(res)
			assert.True(t, dpos.Node(&i).Eq(ex.node))
			assert.Equal(t, dpos.Start(&i), ex.start)
			assert.Equal(t, dpos.End(&i), ex.end)
			if i > 0 {
				b, err := dpos.Before(&i)
				assert.NoError(t, err)
				assert.Equal(t, b, ex.start-1)
				a, err := dpos.After(&i)
				assert.NoError(t, err)
				assert.Equal(t, a, ex.end+1)
			}
		}
		assert.Equal(t, dpos.ParentOffset, exp[len(exp)-3])
		before, err := dpos.NodeBefore()
		assert.NoError(t, err)
		eBefore := exp[len(exp)-2]
		if str, ok := eBefore.(string); ok {
			assert.Equal(t, before.TextContent(), str)
		} else if eBefore == nil {
			assert.Nil(t, before)
		} else {
			assert.Equal(t, before, eBefore)
		}
		after, err := dpos.NodeAfter()
		assert.NoError(t, err)
		eAfter := exp[len(exp)-1]
		if str, ok := eAfter.(string); ok {
			assert.Equal(t, after.TextContent(), str)
		} else if eAfter == nil {
			assert.Nil(t, after)
		} else {
			assert.Equal(t, after, eAfter)
		}
	}
}
