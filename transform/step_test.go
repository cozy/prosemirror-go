package transform

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mkStep(from, to int, val string) Step {
	switch val {
	case "+em":
		return NewAddMarkStep(from, to, schema.Marks["em"].Create(nil))
	case "-em":
		return NewRemoveMarkStep(from, to, schema.Marks["em"].Create(nil))
	default:
		panic(errors.New("Not yet implemented"))
	}
}

func TestStepMerge(t *testing.T) {
	testDoc := doc(p("foobar")).Node

	yes := func(from1, to1 int, val1 string, from2, to2 int, val2 string) {
		step1 := mkStep(from1, to1, val1)
		step2 := mkStep(from2, to2, val2)
		merged, ok := step1.Merge(step2)
		assert.True(t, ok)
		applied1 := step1.Apply(testDoc).Doc
		applied2 := step2.Apply(applied1).Doc
		assert.True(t, merged.Apply(testDoc).Doc.Eq(applied2))
	}

	no := func(from1, to1 int, val1 string, from2, to2 int, val2 string) {
		step1 := mkStep(from1, to1, val1)
		step2 := mkStep(from2, to2, val2)
		_, ok := step1.Merge(step2)
		assert.False(t, ok)
	}

	// TODO
	// // merges typing changes
	// yes(2, 2, "a", 3, 3, "b")

	// // merges inverse typing
	// yes(2, 2, "a", 2, 2, "b")

	// // doesn't merge separated typing
	// no(2, 2, "a", 4, 4, "b")

	// // doesn't merge inverted separated typing
	// no(3, 3, "a", 2, 2, "b")

	// // merges adjacent backspaces
	// yes(3, 4, nil, 2, 3, nil)

	// // merges adjacent deletes
	// yes(2, 3, nil, 2, 3, nil)

	// // doesn't merge separate backspaces
	// no(1, 2, nil, 2, 3, nil)

	// // merges backspace and type
	// yes(2, 3, nil, 2, 2, "x")

	// // merges longer adjacent inserts
	// yes(2, 2, "quux", 6, 6, "baz")

	// // merges inverted longer inserts
	// yes(2, 2, "quux", 2, 2, "baz")

	// // merges longer deletes
	// yes(2, 5, nil, 2, 4, nil)

	// // merges inverted longer deletes
	// yes(4, 6, nil, 2, 4, nil)

	// // merges overwrites
	// yes(3, 4, "x", 4, 5, "y")

	// merges adding adjacent styles
	yes(1, 2, "+em", 2, 4, "+em")

	// merges adding overlapping styles
	yes(1, 3, "+em", 2, 4, "+em")

	// doesn't merge separate styles
	no(1, 2, "+em", 3, 4, "+em")

	// merges removing adjacent styles
	yes(1, 2, "-em", 2, 4, "-em")

	// merges removing overlapping styles
	yes(1, 3, "-em", 2, 4, "-em")

	// doesn't merge removing separate styles
	no(1, 2, "-em", 3, 4, "-em")
}
