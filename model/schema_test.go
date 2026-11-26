package model_test

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/cozy/prosemirror-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaSpecFromJSON(t *testing.T) {
	spec := *schema.Spec
	data, err := json.Marshal(spec)
	assert.NoError(t, err)
	var actual SchemaSpec
	err = json.Unmarshal(data, &actual)
	assert.NoError(t, err)
	assert.Equal(t, spec, actual)
}

func TestCreateAndFillUngeneratableContent(t *testing.T) {
	// 2^40 nodes to generate, without any cycle
	exponential := []*NodeSpec{{Key: "doc", Content: "a0"}, {Key: "text"}}
	for i := 0; i < 40; i++ {
		next := fmt.Sprintf("a%d", i+1)
		exponential = append(exponential, &NodeSpec{Key: fmt.Sprintf("a%d", i), Content: next + " " + next})
	}
	exponential = append(exponential, &NodeSpec{Key: "a40"})

	for _, nodes := range [][]*NodeSpec{
		{{Key: "doc", Content: "a"}, {Key: "a", Content: "a"}, {Key: "text"}},
		{{Key: "doc", Content: "a"}, {Key: "a", Content: "b"}, {Key: "b", Content: "a"}, {Key: "text"}},
		exponential,
	} {
		s, err := NewSchema(&SchemaSpec{Nodes: nodes})
		require.NoError(t, err)
		typ, err := s.NodeType("doc")
		require.NoError(t, err)
		_, err = typ.CreateAndFill()
		assert.Error(t, err)
	}
}
