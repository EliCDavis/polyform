package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	csgSubtractType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/csg.SubtractNode]"
	quadNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/primitives.QuadNode]"
)

func TestInspectionToolsReportAFailedNodeInsideASubgraph(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "drilled", "name": "Drilled",
		"nodes": []map[string]any{
			{"alias": "sheet", "type": quadNodeType, "inputs": map[string]any{
				"Width": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
			}},
			{"alias": "bit", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.6"}}},
			{"alias": "cut", "type": csgSubtractType, "inputs": map[string]any{
				"Base":   map[string]any{"nodeId": "sheet", "port": "Out"},
				"Remove": map[string]any{"elements": []map[string]any{{"nodeId": "bit", "port": "Out"}}},
			}},
		},
		"outputs": []map[string]any{{"name": "Mesh", "nodeId": "cut", "port": "Out"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "drilled"}, &placed)

	var mesh polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": placed.NodeId, "port": "Mesh"}, &mesh)
	assert.Zero(t, mesh.Triangles)
	require.Len(t, mesh.NodeErrors, 1, "%+v", mesh.NodeErrors)
	assert.Equal(t, placed.NodeId+"/", mesh.NodeErrors[0].Path[:len(placed.NodeId)+1])
	assert.Equal(t, "modeling/csg.SubtractNode", mesh.NodeErrors[0].Type)
	assert.Contains(t, mesh.NodeErrors[0].Error, "Base")

	var value polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": placed.NodeId, "port": "Mesh"}, &value)
	assert.Equal(t, mesh.NodeErrors, value.NodeErrors)
}

func TestInspectionToolsReportNothingForAHealthyGraph(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sphereNodeType}, &created)

	var mesh polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": created.NodeId}, &mesh)
	assert.NotZero(t, mesh.Triangles)
	assert.Empty(t, mesh.NodeErrors)
}
