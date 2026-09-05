package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type explodingNode struct{}

func (explodingNode) Out(out *nodes.StructOutput[int]) {
	panic("attribute Position not found")
}

// TestNodePanicNamesTheNode pins that a panic escaping a node's output
// method says which node and port produced it. Nothing recovers between a
// node's method and the caller, so without this a graph-wide failure
// arrives as a bare message with nothing to bisect on.
func TestNodePanicNamesTheNode(t *testing.T) {
	port := nodes.GetNodeOutputPort[int](&nodes.Struct[explodingNode]{}, "Out")

	require.PanicsWithError(t,
		"github.com/EliCDavis/polyform/nodes_test.explodingNode.Out: attribute Position not found",
		func() { port.Value() },
	)
}

// TestNodePanicPreservesErrorText keeps the original message intact so
// callers matching on it still work.
func TestNodePanicPreservesErrorText(t *testing.T) {
	port := nodes.GetNodeOutputPort[int](&nodes.Struct[explodingNode]{}, "Out")

	defer func() {
		r := recover()
		require.NotNil(t, r)
		err, ok := r.(error)
		require.True(t, ok, "panic should carry an error")
		assert.Contains(t, err.Error(), "attribute Position not found")
	}()
	port.Value()
}
