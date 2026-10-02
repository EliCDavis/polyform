package mcp_test

import (
	"strings"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileGraphRollsUpNestedInstances(t *testing.T) {
	session := testSession(t)

	const subtractType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/csg.SubtractNode]"
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "drilled", "name": "Drilled",
		"nodes": []map[string]any{
			{"alias": "block", "type": cubeNodeType, "inputs": map[string]any{
				"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
			}},
			{"alias": "bit", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.6"}}},
			{"alias": "cut", "type": subtractType, "inputs": map[string]any{
				"Base":   map[string]any{"nodeId": "block", "port": "Out"},
				"Remove": map[string]any{"elements": []map[string]any{{"nodeId": "bit", "port": "Out"}}},
			}},
		},
		"outputs": []map[string]any{{"name": "Mesh", "nodeId": "cut", "port": "Out"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "drilled"}, &placed)

	var out polyformmcp.ProfileGraphOutput
	callTool(t, session, "profile_graph", map[string]any{"nodeId": placed.NodeId, "port": "Mesh", "minShare": 0}, &out)

	assert.Equal(t, placed.NodeId+".Mesh", out.Evaluated)
	assert.GreaterOrEqual(t, out.Executions, 3, "block, bit and cut all ran")

	// Which of these three is slowest is decided by sub-microsecond noise
	// on a graph this small, so what is checked is that the time was
	// attributed to all three, not the order it came out in.
	require.NotEmpty(t, out.ByType)
	profiled := make([]string, 0, len(out.ByType))
	for _, entry := range out.ByType {
		profiled = append(profiled, entry.Type)
	}
	for _, want := range []string{"csg.SubtractNode", "primitives.CubeNode", "primitives.UvSphereNode"} {
		assert.Condition(t, func() bool {
			for _, got := range profiled {
				if strings.Contains(got, want) {
					return true
				}
			}
			return false
		}, "%s is missing from %v", want, profiled)
	}

	require.NotEmpty(t, out.Slowest)
	assert.True(t, strings.HasPrefix(out.Slowest[0].Path, placed.NodeId+"/"), "paths run through the instance: %s", out.Slowest[0].Path)

	require.GreaterOrEqual(t, len(out.Flame), 2)
	assert.Contains(t, out.Flame[1], placed.NodeId+" [drilled]")
}

func TestProfileGraphWithoutProducerNeedsANode(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "profile_graph", map[string]any{})
	assert.Contains(t, msg, "no producer")
}
