package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const vec3ValueType = "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector3.Vector[float64]]"

// A subgraph boundary port holds a plain OutputPort, so reflection used to
// accept any output at all. A build wired vector3 literals into a
// []vector3 boundary, was told every connection succeeded, and got a nil
// dereference in MarchNode that named neither the port nor the type - then
// abandoned the tapered curve entirely, which is why that model shipped
// with a bare cone for a tail.
func TestConnectRefusesAMismatchedBoundaryPort(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var tc polyformmcp.CreateTaperedCurveSubgraphOutput
	callTool(t, session, "create_tapered_curve_subgraph", map[string]any{"id": "tail"}, &tc)

	var sub polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail"}, &sub)

	var point polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": vec3ValueType}, &point)

	msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
		"outNodeId": point.NodeId, "outPort": "Value",
		"inNodeId": sub.NodeId, "inPort": "Points",
	})
	assert.Contains(t, msg, "takes []")
	assert.Contains(t, msg, "ArrayFromNodes", "should name the way to build the array")

	// And the port is genuinely still empty rather than holding junk.
	port, ok := inst.Node(sub.NodeId).Inputs()["Points"]
	require.True(t, ok)
	single, ok := port.(nodes.SingleValueInputPort)
	require.True(t, ok, "Points is one array-typed value, not an array port")
	assert.Nil(t, single.Value(), "a refused connection must leave the port untouched")
}

// The matching type is accepted, so the check isn't just refusing things.
func TestConnectAcceptsTheArrayTypeItAsksFor(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var tc polyformmcp.CreateTaperedCurveSubgraphOutput
	callTool(t, session, "create_tapered_curve_subgraph", map[string]any{"id": "tail"}, &tc)

	var sub polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail"}, &sub)

	var arr polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "p0", "type": vec3ValueType},
			map[string]any{"alias": "p1", "type": vec3ValueType},
			map[string]any{
				"alias": "points",
				"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.ArrayFromNodesNode[float64]]",
				"inputs": map[string]any{
					"In": map[string]any{"elements": []any{
						map[string]any{"nodeId": "p0", "port": "Value"},
						map[string]any{"nodeId": "p1", "port": "Value"},
					}},
				},
			},
		},
	}, &arr)
	require.Empty(t, arr.Errors)

	var c polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": arr.Nodes["points"], "outPort": "Out",
		"inNodeId": sub.NodeId, "inPort": "Points",
	}, &c)
	assert.True(t, c.Connected)

	port := inst.Node(sub.NodeId).Inputs()["Points"].(nodes.SingleValueInputPort)
	assert.NotNil(t, port.Value(), "the correctly typed array should have connected")
}

// A subgraph instance has no registered type, so get_node_types can't
// describe it. describe_graph reported only ports that already had a
// connection, which left a freshly placed instance showing its outputs
// and nothing else - and a build guessed at the wiring from there.
func TestDescribeGraphListsUnwiredInputs(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var tc polyformmcp.CreateTaperedCurveSubgraphOutput
	callTool(t, session, "create_tapered_curve_subgraph", map[string]any{"id": "tail"}, &tc)

	var sub polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail"}, &sub)

	var d polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &d)

	var instance *polyformmcp.NodeInstanceSummary
	for i := range d.Nodes {
		if d.Nodes[i].Id == sub.NodeId {
			instance = &d.Nodes[i]
		}
	}
	require.NotNil(t, instance)

	// Every boundary input the tool reported is discoverable, with the
	// type it wants - the []vector3 that caused the trouble included.
	for _, name := range tc.Inputs {
		assert.Contains(t, instance.OpenInputs, name)
	}
	assert.Equal(t, "[]github.com/EliCDavis/vector/vector3.Vector[float64]",
		instance.OpenInputs["Points"])

	// Once wired, a port stops being listed as open.
	var arr polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   "github.com/EliCDavis/polyform/generator/parameter.Value[float64]",
		"inputs": map[string]any{},
	}, &arr)
	var c polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": arr.NodeId, "outPort": "Value",
		"inNodeId": sub.NodeId, "inPort": "Base Radius",
	}, &c)
	require.True(t, c.Connected)

	// A fresh value: unmarshalling into the previous one merges into the
	// existing slice, so an omitted field keeps its old contents.
	var after polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &after)
	for i := range after.Nodes {
		if after.Nodes[i].Id == sub.NodeId {
			assert.NotContains(t, after.Nodes[i].OpenInputs, "Base Radius",
				"a wired port should no longer be listed as open")
			assert.Contains(t, after.Nodes[i].OpenInputs, "Points")
		}
	}
}
