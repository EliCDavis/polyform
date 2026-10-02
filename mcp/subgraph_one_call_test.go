package mcp_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
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

func TestCreateSubgraphOutputTakesItsTypeFromTheFeedingPort(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":   "typed_output_part",
		"name": "Part",
		"nodes": []map[string]any{
			{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
				"Name": map[string]any{"value": `"typed_output_part"`},
			}},
		},
		"outputs": []map[string]any{
			{"name": "Model", "nodeId": "model", "port": "Out"},
		},
	}, &out)
	require.Empty(t, out.Errors)
	require.Contains(t, out.Outputs, "Model")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "typed_output_part"}, &placed)

	port, ok := inst.Node(placed.NodeId).Outputs()["Model"]
	require.True(t, ok)
	assert.Equal(t, "*github.com/EliCDavis/polyform/formats/gltf.PolyformModel", port.(nodes.Typed).Type())
}

func TestCreateSubgraphOutputWithoutTypeOrSourceIsRejected(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	msg := callToolExpectingError(t, session, "create_subgraph", map[string]any{
		"id":      "typed_output_part",
		"name":    "Part",
		"outputs": []map[string]any{{"name": "Model"}},
	})
	assert.Contains(t, msg, "type is required")
}

func TestCreateSubgraphNodesCanInstantiateAHelperSubgraph(t *testing.T) {
	const addNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]"
	session, inst := testSessionWithInstance(t)
	scaledSubgraph(t, session, "scaled")

	// A part whose interior uses the helper twice and adds the results:
	// (Base*2) + (Base*3) = 5*Base, without a single instantiate call.
	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":     "fiver",
		"name":   "Fiver",
		"inputs": []map[string]any{{"name": "Base", "type": "float64"}},
		"nodes": []map[string]any{
			{"alias": "twice", "subgraphId": "scaled", "inputs": map[string]any{
				"Base":  map[string]any{"nodeId": "Base", "port": "Value"},
				"Scale": map[string]any{"value": "2"},
			}},
			{"alias": "thrice", "subgraphId": "scaled", "inputs": map[string]any{
				"Base":  map[string]any{"nodeId": "Base", "port": "Value"},
				"Scale": map[string]any{"value": "3"},
			}},
			{"alias": "sum", "type": addNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"nodeId": "twice", "port": "Result"},
					{"nodeId": "thrice", "port": "Result"},
				}},
			}},
			{"alias": "bad", "type": addNodeType, "subgraphId": "scaled"},
		},
		"outputs": []map[string]any{{"name": "Result", "type": "float64", "nodeId": "sum", "port": "Float"}},
	}, &out)

	require.Len(t, out.Errors, 1, "%v", out.Errors)
	assert.Contains(t, out.Errors[0], "either type or subgraphId")
	require.Contains(t, out.Nodes, "twice")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "fiver",
		"inputs":     map[string]any{"Base": map[string]any{"value": "1.5"}},
	}, &placed)
	assert.InDelta(t, 7.5, evalFloat64Output(t, inst, placed.NodeId, "Result"), 1e-9)
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

	// root: three -> mul(three, four) -> sub(mul, three)
	three := literalFloat64Node(t, session, 3)
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
				"B": map[string]any{"nodeId": three, "port": "Value"},
			}},
		},
	}, &made)
	callTool(t, session, "set_parameter", map[string]any{"nodeId": made.Nodes["four"], "value": "4"}, nil)
	require.InDelta(t, 9, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["four"], made.Nodes["mul"]},
		"name":    "Times Four",
	}, &converted)

	assert.Equal(t, "Times_Four", converted.Id)
	assert.Len(t, converted.Inputs, 1, "'three' is shared with 'sub', so it stays outside and crosses in")
	assert.Len(t, converted.Outputs, 1, "only mul's Float crosses out")
	assert.Empty(t, converted.Literals)
	for _, port := range converted.Inputs {
		assert.Equal(t, three+".Value", port.Connection)
		assert.NotEmpty(t, port.BoundaryNodeId)
	}
	assert.InDelta(t, 9, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9, "downstream must be rewired to the instance")

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{made.Nodes["sub"]}}, &desc)
	for _, n := range desc.Nodes {
		if n.Id == made.Nodes["sub"] {
			assert.Equal(t, converted.NodeId, n.AssignedInput["A"].NodeId)
		}
	}
}

func TestConvertToSubgraphPullsInLiteralsOnlyTheSelectionUses(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []map[string]any{
			{"alias": "mul", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"value": "3"},
					{"value": "4"},
				}},
			}},
			{"alias": "sub", "type": subtractNodeType, "inputs": map[string]any{
				"A": map[string]any{"nodeId": "mul", "port": "Float"},
				"B": map[string]any{"value": "1"},
			}},
		},
	}, &made)
	require.InDelta(t, 11, evalFloat64Output(t, inst, made.Nodes["sub"], "Float"), 1e-9)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{made.Nodes["mul"], made.Nodes["sub"]},
		"name":    "Eleven",
	}, &converted)

	assert.Empty(t, converted.Inputs, "the three literals feed nothing else, so none becomes a port")
	assert.Len(t, converted.Literals, 3)
	assert.Len(t, converted.Outputs, 0, "nothing downstream consumes sub")
	for _, id := range converted.Literals {
		assert.False(t, inst.HasNodeWithId(id), "literal %s should have left the root graph", id)
	}

	child, err := inst.SubGraphInstance(converted.Id)
	require.NoError(t, err)
	var interior int
	for _, id := range child.NodeIds() {
		if _, isParam := child.Node(id).(graph.Parameter); isParam {
			interior++
		}
	}
	assert.Equal(t, 3, interior)
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
		"nodeIds":             []string{made.Nodes["double"], two},
		"name":                "Doubler",
		"keepLiteralsOutside": true,
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

func TestCreateBoundaryNodeNamesExistingInstancesThatNowHaveTheUnwiredPort(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_subgraph", map[string]any{"id": "wheel", "name": "Wheel"}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{"id": "axle", "name": "Axle"}, &polyformmcp.CreateSubgraphOutput{})

	var first polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "wheel", "kind": "input", "portType": "float64", "name": "Radius",
	}, &first)
	assert.Empty(t, first.Instances, "no instances yet")

	var atRoot, nested polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "wheel"}, &atRoot)
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "wheel", "scope": "axle"}, &nested)

	var second polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "wheel", "kind": "input", "portType": "float64", "name": "Width",
	}, &second)
	assert.ElementsMatch(t, []string{"/" + atRoot.NodeId, "axle/" + nested.NodeId}, second.Instances)

	var output polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "wheel", "kind": "output", "portType": "float64", "name": "Out",
	}, &output)
	assert.Empty(t, output.Instances, "an unused output needs no wiring")
}

func TestInstantiateSubgraphCopiesInputsFromASibling(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	scaledSubgraph(t, session, "scaled")
	base := literalFloat64Node(t, session, 3)

	var first polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled",
		"inputs": map[string]any{
			"Base":  map[string]any{"nodeId": base, "port": "Value"},
			"Scale": map[string]any{"value": "4"},
		},
	}, &first)

	var second polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId":     "scaled",
		"copyInputsFrom": first.NodeId,
		"inputs":         map[string]any{"Scale": map[string]any{"value": "10"}},
	}, &second)

	assert.InDelta(t, 12, evalFloat64Output(t, inst, first.NodeId, "Result"), 1e-9)
	assert.InDelta(t, 30, evalFloat64Output(t, inst, second.NodeId, "Result"), 1e-9, "Base shared, Scale overridden")

	callTool(t, session, "set_parameter", map[string]any{"nodeId": base, "value": "5"}, &polyformmcp.SetParameterOutput{})
	assert.InDelta(t, 50, evalFloat64Output(t, inst, second.NodeId, "Result"), 1e-9, "the copied input is the same upstream node, not a snapshot")

	msg := callToolExpectingError(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled", "copyInputsFrom": base,
	})
	assert.Contains(t, msg, "not an instance")
}

func TestConvertToSubgraphMergesReferencesToTheSameVariable(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Scale", "type": "float64", "value": "2"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "a", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"variable": "Scale"}, map[string]any{"value": "3"}}},
		}},
		map[string]any{"alias": "b", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"variable": "Scale"}, map[string]any{"value": "5"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{created.Nodes["a"], created.Nodes["b"]}, "name": "Scaled",
	}, &converted)

	assert.Len(t, converted.Inputs, 1, "two reference nodes to one variable are one port: %v", converted.Inputs)
}

func TestConvertToSubgraphKeepsArrayElementsMixingASourceAndALiteral(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Blend", "type": "float64", "value": "0.5"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "radius", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"variable": "Blend"}, map[string]any{"value": "0.08"}}},
		}},
		map[string]any{"alias": "depth", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"variable": "Blend"}, map[string]any{"value": "-0.16"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{created.Nodes["radius"], created.Nodes["depth"]}, "name": "Rig",
	}, &converted)

	live := inst.Node(converted.NodeId).(*graph.SubgraphInstanceNode).LiveGraph()
	radius := evalFloat64Output(t, live, created.Nodes["radius"], "Float")
	depth := evalFloat64Output(t, live, created.Nodes["depth"], "Float")
	assert.InDelta(t, 0.04, radius, 1e-9, "Blend x 0.08: the literal survives and the variable is read once")
	assert.InDelta(t, -0.08, depth, 1e-9, "Blend x -0.16")
}

func TestConvertToSubgraphKeepsArrayElementsWhenTheLiteralStaysOutside(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	blend := literalFloat64Node(t, session, 0.5)
	coef := literalFloat64Node(t, session, 0.08)

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "radius", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{
				map[string]any{"nodeId": blend, "port": "Value"}, map[string]any{"nodeId": coef, "port": "Value"},
			}},
		}},
		map[string]any{"alias": "depth", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{
				map[string]any{"nodeId": blend, "port": "Value"}, map[string]any{"value": "-0.16"},
			}},
		}},
		// coef is also used outside the selection, so it stays a port.
		map[string]any{"alias": "outside", "type": multiplyNodeType, "inputs": map[string]any{
			"Values": map[string]any{"elements": []any{map[string]any{"nodeId": coef, "port": "Value"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var converted polyformmcp.ConvertToSubgraphOutput
	callTool(t, session, "convert_to_subgraph", map[string]any{
		"nodeIds": []string{created.Nodes["radius"], created.Nodes["depth"]}, "name": "Rig2",
	}, &converted)
	assert.Len(t, converted.Inputs, 1, "only coef is shared with the outside; blend moves in: %v", converted.Inputs)
	assert.Len(t, converted.Literals, 2, "blend and depth's -0.16 move in")

	live := inst.Node(converted.NodeId).(*graph.SubgraphInstanceNode).LiveGraph()
	assert.InDelta(t, 0.04, evalFloat64Output(t, live, created.Nodes["radius"], "Float"), 1e-9)
	assert.InDelta(t, -0.08, evalFloat64Output(t, live, created.Nodes["depth"], "Float"), 1e-9)
}
