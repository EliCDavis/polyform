package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const float64ArrayParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[[]float64]"

func doubleSubgraph(t *testing.T, session *mcpsdk.ClientSession) {
	t.Helper()
	callTool(t, session, "create_equation_subgraph", map[string]any{"id": "double", "equation": "y = x * 2"}, &polyformmcp.CreateEquationSubgraphOutput{})
}

func TestForEachInstanceRunsOncePerItem(t *testing.T) {
	session := testSession(t)
	doubleSubgraph(t, session)

	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "items", "type": float64ArrayParamType, "value": "[1,2,3]"},
		map[string]any{"alias": "each", "subgraphId": "double", "forEach": true, "inputs": map[string]any{
			"x": map[string]any{"nodeId": "items", "port": "Value"},
		}},
	}}, &made)
	require.Empty(t, made.Errors)

	var read polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": made.Nodes["each"], "port": "y"}, &read)
	assert.Equal(t, []any{2.0, 4.0, 6.0}, read.Value)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{made.Nodes["each"]}, "nodeOnly": true}, &desc)
	assert.True(t, desc.Nodes[0].ForEach)
}

func TestSetForEachTogglesAPlacedInstanceAndReportsDroppedEdges(t *testing.T) {
	session := testSession(t)
	doubleSubgraph(t, session)

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "double", "inputs": map[string]any{"x": map[string]any{"value": "2"}}}, &placed)
	var consumer polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   marchNodeType,
		"inputs": map[string]any{"Resolution": map[string]any{"nodeId": placed.NodeId, "port": "y"}},
	}, &consumer)

	var out polyformmcp.SetForEachOutput
	callTool(t, session, "set_for_each", map[string]any{"nodeId": placed.NodeId, "forEach": true}, &out)
	require.Len(t, out.Dropped, 1)
}

func TestInstancePathNamesOneRun(t *testing.T) {
	session := testSession(t)
	doubleSubgraph(t, session)

	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "items", "type": float64ArrayParamType, "value": "[1,2,3]"},
	}}, &made)
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "double", "forEach": true,
		"inputs": map[string]any{"x": map[string]any{"nodeId": made.Nodes["items"], "port": "Value"}},
	}, &placed)

	var subgraphs polyformmcp.ListSubgraphsOutput
	callTool(t, session, "list_subgraphs", map[string]any{}, &subgraphs)
	require.Len(t, subgraphs.Subgraphs, 1)
	assert.Equal(t, []string{placed.NodeId + "[i]"}, subgraphs.Subgraphs[0].Instances)

	var listed polyformmcp.ListNodesOutput
	callTool(t, session, "list_nodes", map[string]any{"scope": "double"}, &listed)
	var multiply string
	for _, n := range listed.Nodes {
		if n.Type == "math.MultiplyNode[float64]" {
			multiply = n.Id
		}
	}
	require.NotEmpty(t, multiply)

	var read polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{
		"nodeId": multiply, "port": "Float", "scope": "double", "instanceId": placed.NodeId + "[1]",
	}, &read)
	assert.Equal(t, 4.0, read.Value)

	msg := callToolExpectingError(t, session, "evaluate_node", map[string]any{
		"nodeId": multiply, "port": "Float", "scope": "double", "instanceId": placed.NodeId,
	})
	assert.Contains(t, msg, "3 run(s)")

	msg = callToolExpectingError(t, session, "evaluate_node", map[string]any{
		"nodeId": multiply, "port": "Float", "scope": "double", "instanceId": placed.NodeId + "[9]",
	})
	assert.Contains(t, msg, "3 run(s)")
}

func TestInstancePathRefusesAnIndexOnAPlainInstance(t *testing.T) {
	session := testSession(t)
	doubleSubgraph(t, session)

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "double"}, &placed)

	msg := callToolExpectingError(t, session, "evaluate_node", map[string]any{
		"nodeId": placed.NodeId, "port": "y", "instanceId": placed.NodeId + "[0]",
	})
	assert.Contains(t, msg, "runs once")
}
