package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chunkNode is the shape every partitioning node wants: an array in, an
// array of arrays out.
type chunkNode struct {
	Values nodes.DynamicPort[[]elemType]
}

func (c chunkNode) Out(out *nodes.Dynamic[[][]elemType]) {
	values, ok := nodes.DynamicArrayValue(c.Values)
	if !ok {
		return
	}

	pairs := make([]any, 0)
	for i := 0; i < values.Len(); i += 2 {
		pairs = append(pairs, values.Range(i, i+2))
	}

	built, ok := nodes.DynamicArrayOf(c.Values, pairs)
	if !ok {
		return
	}
	out.Set(built)
}

func TestDynamicNestedArrayOutputBuildsATypedPort(t *testing.T) {
	node := &nodes.Struct[chunkNode]{}
	connect(t, node, "Values", nodes.ConstOutput[[]string]{Val: []string{"a", "b", "c", "d"}})

	out := node.Outputs()["Out"]
	assert.Equal(t, "[][]string", out.(nodes.Typed).Type())

	typed, ok := out.(nodes.Output[[][]string])
	require.True(t, ok, "a nested array output has to be a real Output[[][]string], got %T", out)
	assert.Equal(t, [][]string{{"a", "b"}, {"c", "d"}}, typed.Value())
}
