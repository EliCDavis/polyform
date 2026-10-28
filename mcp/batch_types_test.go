package mcp_test

import (
	"strings"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stringArrayParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[[]string]"

func TestCreateNodesChecksAnEdgeAgainstTypesTheWholeBatchSettles(t *testing.T) {
	session := testSession(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "stroke", "type": strokeNodeType, "inputs": map[string]any{
			"Path": map[string]any{"nodeId": "point", "port": "Out"},
		}},
		map[string]any{"alias": "point", "type": vector2NewNodeType, "inputs": map[string]any{
			"X": map[string]any{"nodeId": "xs", "port": "Value"},
			"Y": map[string]any{"value": "1"},
		}},
		map[string]any{"alias": "xs", "type": floatArrayParamType, "value": "[1,2,3]"},
	}}, &made)

	assert.Empty(t, made.Errors)
}

func TestCreateNodesTypesALiteralElementFromAReferenceLaterInTheList(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "names", "type": stringArrayParamType, "value": `["c"]`},
		map[string]any{"alias": "pick", "type": switchNodeType, "inputs": map[string]any{
			"Index": map[string]any{"value": "0"},
			"Cases": map[string]any{"elements": []any{
				map[string]any{"value": `["a"]`},
				map[string]any{"value": `["b"]`},
				map[string]any{"nodeId": "names", "port": "Value"},
			}},
		}},
	}}, &made)
	require.Empty(t, made.Errors)

	cases := inst.NodeInstanceSchema(inst.Node(made.Nodes["pick"])).AssignedInput
	assert.Equal(t, made.Nodes["names"], cases["Cases.2"].NodeId, "each element keeps its place")
}

func TestCreateNodesWiresNoElementOfAPortWhenOneFails(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "three", "type": floatParamType, "value": "3"},
		map[string]any{"alias": "sum", "type": sumNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{
				map[string]any{"nodeId": "missing", "port": "Value"},
				map[string]any{"nodeId": "three", "port": "Value"},
			}},
		}},
	}}, &made)

	require.NotEmpty(t, made.Errors)
	assert.Empty(t, inst.NodeInstanceSchema(inst.Node(made.Nodes["sum"])).AssignedInput, "a later element would have slid into the failed one's place")
}

func TestCreateSubgraphKeepsTheSubgraphWhenAnInteriorEdgeIsDropped(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":   "joined",
		"name": "Joined",
		"nodes": []any{
			map[string]any{"alias": "one", "type": floatParamType, "value": "1"},
			map[string]any{"alias": "cat", "type": concatNodeType, "inputs": map[string]any{
				"Arrays": map[string]any{"elements": []any{map[string]any{"nodeId": "one", "port": "Value"}}},
			}},
		},
	}, &out)

	_, err := inst.SubGraphInstance("joined")
	require.NoError(t, err)
	assert.NotEmpty(t, out.Nodes["cat"])
	errs := strings.Join(out.Errors, "\n")
	assert.Contains(t, errs, "removed")
	assert.Contains(t, errs, "cat", "the edge is named by the batch's aliases")
}

func TestConvertToSubgraphRefusalDoesNotClaimEdgesWereRemoved(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "values", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "spread", "type": spreadNodeType, "inputs": map[string]any{
			"Array": map[string]any{"nodeId": "values", "port": "Value"},
		}},
		map[string]any{"alias": "sum", "type": sumNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	}}, &made)
	require.Empty(t, made.Errors)

	msg := callToolExpectingError(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["sum"]},
		"name":    "Summed",
	})
	assert.NotContains(t, msg, "removed")
	assert.True(t, inst.HasNodeWithId(made.Nodes["sum"]), "the conversion was rolled back")
	assert.InDelta(t, 6, evalFloat64Output(t, inst, made.Nodes["sum"], "Float"), 1e-9)
}
