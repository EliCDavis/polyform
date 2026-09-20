package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A "scaled" part: Value = Base * Scale, with both as boundary inputs.
func scaledSubgraph(t *testing.T, session *mcpsdk.ClientSession, id string) polyformmcp.CreateSubgraphOutput {
	t.Helper()
	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":   id,
		"name": "Scaled",
		"inputs": []map[string]any{
			{"name": "Base", "type": "float64"},
			{"name": "Scale", "type": "float64"},
		},
		"nodes": []map[string]any{
			{"alias": "product", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"nodeId": "Base", "port": "Value"},
					{"nodeId": "Scale", "port": "Value"},
				}},
			}},
		},
		"outputs": []map[string]any{
			{"name": "Result", "type": "float64", "nodeId": "product", "port": "Float"},
		},
	}, &out)
	return out
}

func TestCreateSubgraphInOneCall(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	out := scaledSubgraph(t, session, "scaled")

	require.Empty(t, out.Errors)
	assert.Equal(t, "scaled", out.Id)
	assert.Len(t, out.Inputs, 2)
	assert.Len(t, out.Outputs, 1)
	assert.Contains(t, out.Nodes, "product")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled",
		"inputs": map[string]any{
			"Base":  map[string]any{"value": "3"},
			"Scale": map[string]any{"value": "4"},
		},
	}, &placed)

	assert.InDelta(t, 12, evalFloat64Output(t, inst, placed.NodeId, "Result"), 1e-9)
}

func TestCreateSubgraphRejectsAliasShadowingAnInput(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "create_subgraph", map[string]any{
		"id": "bad", "name": "Bad",
		"inputs": []map[string]any{{"name": "Base", "type": "float64"}},
		"nodes":  []map[string]any{{"alias": "Base", "type": floatParamType}},
	})
	assert.Contains(t, msg, "boundary port")

	var list polyformmcp.ListSubgraphsOutput
	callTool(t, session, "list_subgraphs", map[string]any{}, &list)
	assert.Empty(t, list.Subgraphs, "a failed create must not leave a half-built definition behind")
}

func TestCreateSubgraphReportsBadInteriorEntryButKeepsTheRest(t *testing.T) {
	session := testSession(t)
	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "partial", "name": "Partial",
		"inputs": []map[string]any{{"name": "Base", "type": "float64"}},
		"nodes": []map[string]any{
			{"alias": "ok", "type": floatParamType},
			{"alias": "broken", "type": "no/such.Type"},
		},
	}, &out)

	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "broken")
	assert.Contains(t, out.Nodes, "ok")
	assert.Len(t, out.Inputs, 1)
}

func TestConvertToSubgraphKeepsTheGraphComputingTheSameThing(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	// root: three -> mul(three, four) -> sub(mul, one)
	three := literalFloat64Node(t, session, 3)
	one := literalFloat64Node(t, session, 1)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []map[string]any{
			{"alias": "four", "type": floatParamType},
			{"alias": "mul", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"nodeId": three, "port": "Value"},
					{"nodeId": "four", "port": "Value"},
				}},
			}},
			{"alias": "sub", "type": subtractNodeType, "inputs": map[string]any{
				"A": map[string]any{"nodeId": "mul", "port": "Float"},
				"B": map[string]any{"nodeId": one, "port": "Value"},
			}},
		},
	}, &made)
	callTool(t, session, "set_parameter", map[string]any{"nodeId": made.Nodes["four"], "value": "4"}, nil)
	require.InDelta(t, 11, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["four"], made.Nodes["mul"]},
		"name":    "Times Four",
	}, &converted)

	assert.Equal(t, "Times_Four", converted.Id)
	assert.Len(t, converted.Inputs, 1, "only 'three' crosses in")
	assert.Len(t, converted.Outputs, 1, "only mul's Float crosses out")
	for _, port := range converted.Inputs {
		assert.Equal(t, three+".Value", port.Connection)
		assert.NotEmpty(t, port.BoundaryNodeId)
	}
	assert.InDelta(t, 11, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9, "downstream must be rewired to the instance")

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{made.Nodes["sub"]}}, &desc)
	for _, n := range desc.Nodes {
		if n.Id == made.Nodes["sub"] {
			assert.Equal(t, converted.NodeId, n.AssignedInput["A"].NodeId)
		}
	}
}

func TestRenameBoundaryPortFollowsOnInstances(t *testing.T) {
	session := testSession(t)
	three := literalFloat64Node(t, session, 3)
	two := literalFloat64Node(t, session, 2)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []map[string]any{
			{"alias": "double", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"nodeId": three, "port": "Value"},
					{"nodeId": two, "port": "Value"},
				}},
			}},
		},
	}, &made)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["double"], two},
		"name":    "Doubler",
	}, &converted)

	var inputName string
	var inputBoundary string
	for name, port := range converted.Inputs {
		inputName, inputBoundary = name, port.BoundaryNodeId
	}
	require.Equal(t, "Input 1", inputName)

	callTool(t, session, "rename_boundary_port", map[string]any{
		"subgraphId": converted.Id, "nodeId": inputBoundary, "name": "In",
	}, &polyformmcp.RenameBoundaryPortOutput{})

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{converted.NodeId}}, &desc)
	for _, n := range desc.Nodes {
		if n.Id == converted.NodeId {
			ref, hasIn := n.AssignedInput["In"]
			require.True(t, hasIn, "instance should show the port under its new name; has %v", n.AssignedInput)
			assert.Equal(t, three, ref.NodeId, "wiring through the renamed port survives")
		}
	}
}
