package debug_test

import (
	"testing"

	"github.com/EliCDavis/polyform/debug"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wire(t *testing.T, node nodes.Node, port string, source nodes.OutputPort) {
	t.Helper()
	input, ok := node.Inputs()[port].(nodes.SingleValueInputPort)
	require.True(t, ok, "node has no single-value input %q", port)
	require.NoError(t, input.Set(source))
}

func TestProbePassesTheValueThroughUntouched(t *testing.T) {
	cube := nodes.GetNodeOutputPort[modeling.Mesh](&nodes.Struct[primitives.CubeNode]{}, "Out")

	node := &nodes.Struct[debug.ProbeNode]{}
	wire(t, node, "In", cube)

	out := node.Outputs()["Out"]
	typed, ok := out.(nodes.Output[modeling.Mesh])
	require.True(t, ok, "a probe of a mesh really is an Output[Mesh], got %T", out)
	assert.Equal(t, cube.Value().PrimitiveCount(), typed.Value().PrimitiveCount())
}

func TestProbeReportsTheTypeItIsCarrying(t *testing.T) {
	node := &nodes.Struct[debug.ProbeNode]{}
	wire(t, node, "In", nodes.ConstOutput[float64]{Val: 3})

	assert.Equal(t, "float64", nodes.GetNodeOutputPort[string](node, "Type").Value())
	assert.Equal(t, "3", nodes.GetNodeOutputPort[string](node, "Value").Value())
	assert.True(t, nodes.GetNodeOutputPort[bool](node, "Connected").Value())
}

func TestProbeWithNothingWiredIn(t *testing.T) {
	node := &nodes.Struct[debug.ProbeNode]{}

	assert.Equal(t, "", nodes.GetNodeOutputPort[string](node, "Type").Value())
	assert.Equal(t, "", nodes.GetNodeOutputPort[string](node, "Value").Value())
	assert.False(t, nodes.GetNodeOutputPort[bool](node, "Connected").Value())
	assert.Equal(t, "", node.Outputs()["Out"].(nodes.Typed).Type(),
		"an unprobed pass-through has no type, so anything may still connect")
}

func TestProbeOutputsCarryTheirDescriptions(t *testing.T) {
	node := &nodes.Struct[debug.ProbeNode]{}
	wire(t, node, "In", nodes.ConstOutput[float64]{Val: 3})

	describable, ok := node.Outputs()["Out"].(nodes.Describable)
	require.True(t, ok, "a bound dynamic output still has to carry its description")
	assert.Equal(t, "The value being looked at, passed through unchanged.", describable.Description())
}
