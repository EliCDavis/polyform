package mcp_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/EliCDavis/polyform/generator/subgraph"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// root: three -> mul(three, four) -> sub(mul, three), computing 9.
func timesFourGraph(t *testing.T, session *mcpsdk.ClientSession) (three string, made polyformmcp.CreateNodesOutput) {
	t.Helper()
	three = literalFloat64Node(t, session, 3)
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
				"B": map[string]any{"nodeId": three, "port": "Value"},
			}},
		},
	}, &made)
	callTool(t, session, "set_parameter", map[string]any{"nodeId": made.Nodes["four"], "value": "4"}, nil)
	return three, made
}

func TestConvertToSubgraphPortsMapSetsTheFinalNamesInTheSameCall(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	three, made := timesFourGraph(t, session)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["four"], made.Nodes["mul"]},
		"name":    "Times Four",
		"ports": map[string]string{
			"Value":                      "Base",
			made.Nodes["mul"] + ".Float": "Product",
		},
	}, &converted)

	require.Equal(t, []string{"Base"}, slices.Sorted(maps.Keys(converted.Inputs)))
	assert.Equal(t, three+".Value", converted.Inputs["Base"].Connection)
	require.Equal(t, []string{"Product"}, slices.Sorted(maps.Keys(converted.Outputs)))
	assert.Equal(t, made.Nodes["mul"]+".Float", converted.Outputs["Product"].Connection)
	assert.InDelta(t, 9, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9)

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{made.Nodes["sub"]}}, &desc)
	for _, n := range desc.Nodes {
		if n.Id == made.Nodes["sub"] {
			assert.Equal(t, converted.NodeId, n.AssignedInput["A"].NodeId)
			assert.Equal(t, "Product", n.AssignedInput["A"].Port)
		}
	}
}

func TestConvertToSubgraphPortsMapWithAnUnknownKeyChangesNothing(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	_, made := timesFourGraph(t, session)

	msg := callToolExpectingError(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["four"], made.Nodes["mul"]},
		"name":    "Times Four",
		"ports":   map[string]string{"Nope": "Base"},
	})
	assert.Contains(t, msg, `"Nope"`)
	assert.Contains(t, msg, `"Value"`, "the generated names are listed for the retry")

	assert.True(t, inst.HasNodeWithId(made.Nodes["mul"]), "the conversion was rolled back")
	_, err := inst.SubGraphInstance("Times_Four")
	assert.Error(t, err)
	assert.InDelta(t, 9, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9)
}

func TestConvertToSubgraphNamesAnInputAfterTheSubgraphInstanceFeedingIt(t *testing.T) {
	session := testSession(t)
	scaledSubgraph(t, session, "scaled")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled",
		"inputs": map[string]any{
			"Base":  map[string]any{"value": "3"},
			"Scale": map[string]any{"value": "2"},
		},
	}, &placed)

	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []map[string]any{
			{"alias": "mul", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"nodeId": placed.NodeId, "port": "Result"},
					{"value": "2"},
				}},
			}},
		},
	}, &made)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["mul"]},
		"name":    "Part",
	}, &converted)

	assert.Equal(t, []string{"Scaled"}, slices.Sorted(maps.Keys(converted.Inputs)))
}

func TestConvertToSubgraphNamesAnOutputAfterTheModelNodesName(t *testing.T) {
	session := testSession(t)

	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)
	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfModelType,
		"inputs": map[string]any{
			"Name": map[string]any{"value": `"Body"`},
			"Mesh": map[string]any{"nodeId": cube.NodeId, "port": "Out"},
		},
	}, &model)
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": model.NodeId, "port": "Out"}}},
		},
	}, &polyformmcp.CreateNodeOutput{})

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{model.NodeId},
		"name":    "Body Part",
	}, &converted)

	assert.Equal(t, []string{"Body"}, slices.Sorted(maps.Keys(converted.Outputs)))
	assert.Equal(t, []string{"Mesh"}, slices.Sorted(maps.Keys(converted.Inputs)), "a generic Out falls back to the input it feeds")
	assert.Len(t, converted.Literals, 1, "the Name literal moves in with the model")
}

func TestRenameBoundaryPortBatchLetsNamesTradePlaces(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	made := scaledSubgraph(t, session, "scaled")

	var out polyformmcp.RenameBoundaryPortOutput
	callTool(t, session, "rename_boundary_port", map[string]any{
		"subgraphId": "scaled",
		"renames": []map[string]any{
			{"nodeId": made.Inputs["Base"], "name": "Scale"},
			{"port": "Scale", "name": "Base"},
		},
	}, &out)
	assert.True(t, out.Renamed)
	assert.Equal(t, 2, out.Count)

	child, err := inst.SubGraphInstance("scaled")
	require.NoError(t, err)
	assert.Equal(t, "Scale", child.Node(made.Inputs["Base"]).(*subgraph.InputNode).PortName)
	assert.Equal(t, "Base", child.Node(made.Inputs["Scale"]).(*subgraph.InputNode).PortName)
}

func TestRenameBoundaryPortBatchFailsAsAWhole(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	made := scaledSubgraph(t, session, "scaled")

	msg := callToolExpectingError(t, session, "rename_boundary_port", map[string]any{
		"subgraphId": "scaled",
		"renames": []map[string]any{
			{"nodeId": made.Inputs["Base"], "name": "Width"},
			{"port": "Nope", "name": "Height"},
		},
	})
	assert.Contains(t, msg, "renames[1]")
	assert.Contains(t, msg, `"Nope"`)

	child, err := inst.SubGraphInstance("scaled")
	require.NoError(t, err)
	assert.Equal(t, "Base", child.Node(made.Inputs["Base"]).(*subgraph.InputNode).PortName)
}
