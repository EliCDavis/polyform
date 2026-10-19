package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenameSubgraphMovesInstancesAndOutlinePart(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "set_outline", map[string]any{"yaml": dogOutline}, &polyformmcp.SetOutlineOutput{})
	scaledSubgraph(t, session, "tail")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "tail",
		"inputs":     map[string]any{"Base": map[string]any{"value": "3"}, "Scale": map[string]any{"value": "4"}},
	}, &placed)

	var out polyformmcp.RenameSubgraphOutput
	callTool(t, session, "rename_subgraph", map[string]any{"id": "tail", "newId": "tail_curved"}, &out)
	assert.Equal(t, "tail_curved", out.Id)
	assert.Equal(t, []string{"/" + placed.NodeId}, out.Instances)
	assert.True(t, out.OutlineRenamed)

	var subgraphs polyformmcp.ListSubgraphsOutput
	callTool(t, session, "list_subgraphs", map[string]any{}, &subgraphs)
	require.Len(t, subgraphs.Subgraphs, 1)
	assert.Equal(t, "tail_curved", subgraphs.Subgraphs[0].Id)
	assert.Equal(t, []string{placed.NodeId}, subgraphs.Subgraphs[0].Instances)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{placed.NodeId}, "nodeOnly": true}, &desc)
	assert.Equal(t, "tail_curved", desc.Nodes[0].SubGraphId)
	assert.InDelta(t, 12, evalFloat64Output(t, inst, placed.NodeId, "Result"), 1e-9)

	var outline polyformmcp.GetOutlineOutput
	callTool(t, session, "get_outline", map[string]any{}, &outline)
	assert.Contains(t, outline.Yaml, "tail_curved")
	assert.NotContains(t, outline.Yaml, "tail:")

	var again polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail_curved"}, &again)
	callToolExpectingError(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "tail"})
}

func TestRenameSubgraphRefusesATakenId(t *testing.T) {
	session := testSession(t)
	scaledSubgraph(t, session, "a")
	scaledSubgraph(t, session, "b")
	msg := callToolExpectingError(t, session, "rename_subgraph", map[string]any{"id": "a", "newId": "b"})
	assert.Contains(t, msg, "already exists")
}

func TestListNodesIsIdTypeAndNameOnly(t *testing.T) {
	session := testSession(t)
	scaledSubgraph(t, session, "scaled")
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "scaled"}, &placed)

	var root polyformmcp.ListNodesOutput
	callTool(t, session, "list_nodes", map[string]any{}, &root)
	require.Len(t, root.Nodes, 1)
	assert.Equal(t, placed.NodeId, root.Nodes[0].Id)
	assert.Equal(t, "scaled", root.Nodes[0].SubgraphId)
	assert.Equal(t, "Scaled", root.Nodes[0].Name)

	var inside polyformmcp.ListNodesOutput
	callTool(t, session, "list_nodes", map[string]any{"scope": "scaled"}, &inside)
	assert.Equal(t, "scaled", inside.Scope)
	assert.Len(t, inside.Nodes, 4, "two inputs, the multiply, one output")
	var types []string
	for _, n := range inside.Nodes {
		types = append(types, n.Type)
	}
	assert.Contains(t, types, "math.MultiplyNode[float64]")
}
