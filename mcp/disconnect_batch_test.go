package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisconnectBatchResolvesIndicesBeforeRemoving(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	fourth := literalFloat64Node(t, session, 40)

	var sum polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]",
		"inputs": map[string]any{"Values": map[string]any{"elements": []map[string]any{
			{"value": "1"}, {"value": "2"}, {"value": "3"}, {"nodeId": fourth, "port": "Value"},
		}}},
	}, &sum)
	require.InDelta(t, 46, evalFloat64Output(t, inst, sum.NodeId, "Float"), 1e-9)

	var out polyformmcp.DisconnectOutput
	callTool(t, session, "disconnect", map[string]any{"connections": []map[string]any{
		{"nodeId": sum.NodeId, "port": "Values.0"},
		{"nodeId": sum.NodeId, "port": "Values.2"},
		{"nodeId": sum.NodeId, "port": "Values.@" + fourth},
	}}, &out)

	assert.True(t, out.Disconnected)
	assert.Equal(t, 3, out.Removed)
	require.Len(t, out.Arrays, 1)
	assert.Equal(t, sum.NodeId+".Values", out.Arrays[0].Port)
	assert.Len(t, out.Arrays[0].Remaining, 1)
	assert.InDelta(t, 2, evalFloat64Output(t, inst, sum.NodeId, "Float"), 1e-9, "only the second element is left")
}

func TestDisconnectBatchRemovesNothingWhenAnEntryIsBad(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var sum polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]",
		"inputs": map[string]any{"Values": map[string]any{"elements": []map[string]any{
			{"value": "1"}, {"value": "2"},
		}}},
	}, &sum)

	msg := callToolExpectingError(t, session, "disconnect", map[string]any{"connections": []map[string]any{
		{"nodeId": sum.NodeId, "port": "Values.0"},
		{"nodeId": sum.NodeId, "port": "Values.9"},
	}})
	assert.Contains(t, msg, "connection 1")
	assert.InDelta(t, 3, evalFloat64Output(t, inst, sum.NodeId, "Float"), 1e-9)
}

func TestDisconnectSingleStillReportsRemaining(t *testing.T) {
	session := testSession(t)
	var sum polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]",
		"inputs": map[string]any{"Values": map[string]any{"elements": []map[string]any{
			{"value": "1"}, {"value": "2"},
		}}},
	}, &sum)

	var out polyformmcp.DisconnectOutput
	callTool(t, session, "disconnect", map[string]any{"nodeId": sum.NodeId, "port": "Values.0"}, &out)
	assert.True(t, out.Disconnected)
	assert.Len(t, out.Remaining, 1)
	assert.Empty(t, out.Arrays)
}
