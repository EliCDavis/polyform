package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteNodeBatchKeepsGoingPastABadId(t *testing.T) {
	session := testSession(t)
	first := literalFloat64Node(t, session, 1)
	second := literalFloat64Node(t, session, 2)

	var out polyformmcp.DeleteNodeOutput
	callTool(t, session, "delete_node", map[string]any{"nodes": []any{
		map[string]any{"nodeId": first},
		map[string]any{"nodeId": "not-a-real-node"},
		map[string]any{"nodeId": second},
	}}, &out)

	assert.Equal(t, 2, out.Deleted)
	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "node 1", "the error should name which entry failed")

	var after polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{}, &after)
	for _, n := range after.Nodes {
		assert.NotEqual(t, first, n.Id, "a good id in the batch should still be deleted")
		assert.NotEqual(t, second, n.Id, "a good id in the batch should still be deleted")
	}
}

func TestDeleteNodeSingleBadIdIsStillAToolError(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "delete_node", map[string]any{"nodeId": "not-a-real-node"})
	assert.Contains(t, msg, "not-a-real-node")
}

func TestDeleteNodeListsEachOrphanedElementOnce(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "mul", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{
				map[string]any{"value": "2"},
				map[string]any{"value": "3"},
			}},
		}},
	}}, &created)
	mul := created.Nodes["mul"]
	zero := created.Literals["mul.Values.0"]
	one := created.Literals["mul.Values.1"]
	require.NotEmpty(t, zero)
	require.NotEmpty(t, one)

	var out polyformmcp.DeleteNodeOutput
	callTool(t, session, "delete_node", map[string]any{"nodes": []any{
		map[string]any{"nodeId": zero},
		map[string]any{"nodeId": one},
	}}, &out)

	assert.Equal(t, 2, out.Deleted)
	assert.Equal(t, []string{mul + ".Values.0", mul + ".Values.1"}, out.Orphaned,
		"both elements went, at the indices they held before the call")
}

func TestDeleteNodeOmitsPortsOnNodesItAlsoDeleted(t *testing.T) {
	session := testSession(t)
	three := literalFloat64Node(t, session, 3)
	mul := multiplyOf(t, session, "2")
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": three, "outPort": "Value", "inNodeId": mul, "inPort": "Values",
	}, &polyformmcp.ConnectNodesOutput{})

	var out polyformmcp.DeleteNodeOutput
	callTool(t, session, "delete_node", map[string]any{"nodes": []any{
		map[string]any{"nodeId": three},
		map[string]any{"nodeId": mul},
	}}, &out)

	assert.Equal(t, 2, out.Deleted)
	assert.Empty(t, out.Orphaned, "the only consumer went in the same call")
}
