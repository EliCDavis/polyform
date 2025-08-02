package nodes_test

import (
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

func TestGetNodeOutputPort(t *testing.T) {
	val1 := 123.456
	val2 := 678.910
	node := &nodes.Struct[SimpleAddTestStructNode]{
		Data: SimpleAddTestStructNode{
			A: nodes.ConstOutput[float64]{Val: val1},
			B: nodes.ConstOutput[float64]{Val: val2},
		},
	}

	// ACT ====================================================================
	port := nodes.GetNodeOutputPort[float64](node, "Sum")

	// ASSERT =================================================================
	assert.Equal(t, val1+val2, port.Value())

	assert.PanicsWithError(t, "node port \"Sum\" is not type string", func() {
		nodes.GetNodeOutputPort[string](node, "Sum")
	})

	assert.PanicsWithError(t, "node does not contain a port named \"Wrong\", only Sum ", func() {
		nodes.GetNodeOutputPort[string](node, "Wrong")
	})
}

func TestTryGetOutputValue(t *testing.T) {
	out := &nodes.StructOutput[string]{}
	assert.Equal(t, 123., nodes.TryGetOutputValue(out, nodes.ConstOutput[float64]{Val: 123}, 456))
	assert.Equal(t, 456., nodes.TryGetOutputValue[float64](out, nil, 456))
}
