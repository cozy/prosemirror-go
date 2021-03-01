package model_test

import (
	"encoding/json"
	"testing"

	. "github.com/shodgson/prosemirror-go/model"
	"github.com/stretchr/testify/assert"
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
