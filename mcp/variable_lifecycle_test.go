package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scale is read by a multiply inside "inner", which "outer" places, which
// the root places: the reference sits two subgraph levels down.
func nestedVariableReader(t *testing.T) (session sessionT, outerInstance string) {
	t.Helper()
	session = testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Scale", "type": "float64", "value": "2"}},
	}, &polyformmcp.CreateVariablesOutput{})

	callTool(t, session, "create_subgraph", map[string]any{
		"id": "inner", "name": "Inner",
		"nodes": []map[string]any{
			{"alias": "mul", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []any{map[string]any{"variable": "Scale"}, map[string]any{"value": "3"}}},
			}},
		},
		"outputs": []map[string]any{{"name": "Out", "nodeId": "mul", "port": "Float"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	callTool(t, session, "create_subgraph", map[string]any{"id": "outer", "name": "Outer"}, &polyformmcp.CreateSubgraphOutput{})
	var innerInOuter polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "inner", "scope": "outer"}, &innerInOuter)
	var boundary polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "outer", "kind": "output", "portType": "float64", "name": "Out",
	}, &boundary)
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": innerInOuter.NodeId, "outPort": "Out",
		"inNodeId": boundary.NodeId, "inPort": "Value", "scope": "outer",
	}, &polyformmcp.ConnectNodesOutput{})

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "outer"}, &placed)
	return session, placed.NodeId
}

type sessionT = *mcpsdk.ClientSession

func evaluateFloat(t *testing.T, session sessionT, nodeID, port string) float64 {
	t.Helper()
	var out polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": nodeID, "port": port}, &out)
	value, ok := out.Value.(float64)
	require.True(t, ok, "%s.%s evaluated to %#v, not a number", nodeID, port, out.Value)
	return value
}

func TestRenameVariableReachesReferencesTwoSubgraphsDeep(t *testing.T) {
	session, placed := nestedVariableReader(t)
	require.InDelta(t, 6.0, evaluateFloat(t, session, placed, "Out"), 1e-9)

	var renamed polyformmcp.RenameVariableOutput
	callTool(t, session, "rename_variable", map[string]any{"path": "Scale", "newPath": "Body/Scale"}, &renamed)
	callTool(t, session, "update_variable", map[string]any{"path": "Body/Scale", "value": "5"}, &polyformmcp.UpdateVariableOutput{})

	assert.InDelta(t, 15.0, evaluateFloat(t, session, placed, "Out"), 1e-9, "the placement built before the rename")

	var fresh polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "outer"}, &fresh)
	assert.InDelta(t, 15.0, evaluateFloat(t, session, fresh.NodeId, "Out"), 1e-9, "a placement built after the rename, two levels deep")

	var shallow polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "inner"}, &shallow)
	assert.InDelta(t, 15.0, evaluateFloat(t, session, shallow.NodeId, "Out"), 1e-9, "a placement built after the rename, one level deep")

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	assert.Equal(t, "Body/Scale", list.Variables[0].Path)
	require.Len(t, list.Variables[0].References, 1)
	assert.Contains(t, list.Variables[0].References[0], "inner/")
}

func TestRenameVariableRewritesProfilesAndVariantSets(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": "1"},
			{"path": "Count", "type": "int", "value": "2"},
		},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Wide", "variables": map[string]any{"Width": "4"}}, &polyformmcp.SaveProfileOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "Widths",
		"dimensions": []map[string]any{{"path": "Width", "type": "numericRange", "data": `{"min":0,"max":1,"samples":3}`}},
	}, &polyformmcp.CreateVariantSetOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Looks",
		"dimensions": []map[string]any{
			{"path": "Width", "type": "discrete", "data": `{"values":[2]}`},
			{"path": "Count", "type": "discrete", "data": `{"values":[3]}`},
		},
	}, &polyformmcp.CreateVariantSetOutput{})

	var renamed polyformmcp.RenameVariableOutput
	callTool(t, session, "rename_variable", map[string]any{"path": "Width", "newPath": "Frame Width"}, &renamed)
	assert.Equal(t, []string{"Wide"}, renamed.Profiles)
	assert.Equal(t, []string{"Looks", "Widths"}, renamed.VariantSets)

	var profile polyformmcp.GetProfileOutput
	callTool(t, session, "get_profile", map[string]any{"name": "Wide"}, &profile)
	assert.Equal(t, map[string]string{"Frame Width": "4", "Count": "2"}, profile.Variables)
	assert.Empty(t, profile.Stale)

	var sets polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &sets)
	require.Len(t, sets.VariantSets, 2)
	assert.Equal(t, "Frame Width", sets.VariantSets[1].Dimensions[0].Path)

	var run polyformmcp.RunVariantSweepOutput
	callTool(t, session, "run_variant_sweep", map[string]any{"name": "Looks", "outputDir": t.TempDir()}, &run)
	require.Len(t, run.Variants, 1)
	assert.Equal(t, map[string]string{"Frame Width": "2", "Count": "3"}, run.Variants[0].Variables)
}

func TestDeleteVariableRemovesEveryReferenceAndProfileKey(t *testing.T) {
	session, placed := nestedVariableReader(t)
	require.InDelta(t, 6.0, evaluateFloat(t, session, placed, "Out"), 1e-9)

	var orphanRoot, orphanInner polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": "Scale"}, &orphanRoot)
	callTool(t, session, "create_node", map[string]any{"type": "Scale", "scope": "inner"}, &orphanInner)

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Count", "type": "int", "value": "2"}},
	}, &polyformmcp.CreateVariablesOutput{})
	callTool(t, session, "save_profile", map[string]any{"name": "Big", "variables": map[string]any{"Scale": "9"}}, &polyformmcp.SaveProfileOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name":       "ScaleOnly",
		"dimensions": []map[string]any{{"path": "Scale", "type": "numericRange", "data": `{"min":0,"max":1,"samples":3}`}},
	}, &polyformmcp.CreateVariantSetOutput{})
	callTool(t, session, "create_variant_set", map[string]any{
		"name": "Both",
		"dimensions": []map[string]any{
			{"path": "Scale", "type": "numericRange", "data": `{"min":0,"max":1,"samples":3}`},
			{"path": "Count", "type": "intRange", "data": `{"min":1,"max":2,"samples":2}`},
		},
	}, &polyformmcp.CreateVariantSetOutput{})

	var deleted polyformmcp.DeleteVariableOutput
	callTool(t, session, "delete_variable", map[string]any{"path": "Scale"}, &deleted)
	assert.True(t, deleted.Deleted)
	require.Len(t, deleted.References, 3, "%v", deleted.References)
	assert.Contains(t, deleted.References, orphanRoot.NodeId)
	assert.Contains(t, deleted.References, "inner/"+orphanInner.NodeId)
	require.Len(t, deleted.Unfed, 1, "%v", deleted.Unfed)
	assert.Contains(t, deleted.Unfed[0], "inner/")
	assert.Contains(t, deleted.Unfed[0], ".Values")
	assert.Equal(t, []string{"Big"}, deleted.Profiles)
	assert.Equal(t, []string{"Both"}, deleted.VariantSets)
	assert.Equal(t, []string{"ScaleOnly"}, deleted.VariantSetsDeleted)

	for _, scope := range []string{"", "inner", "outer"} {
		var desc polyformmcp.DescribeGraphOutput
		callTool(t, session, "describe_graph", map[string]any{"scope": scope}, &desc)
		for _, node := range desc.Nodes {
			assert.NotEqual(t, "Scale", node.Type, "scope %q still holds a reference node", scope)
		}
	}
	assert.InDelta(t, 3.0, evaluateFloat(t, session, placed, "Out"), 1e-9, "the multiply still runs on its remaining element")

	var profiles polyformmcp.ListProfilesOutput
	callTool(t, session, "list_profiles", map[string]any{}, &profiles)
	require.Len(t, profiles.Profiles, 1)
	assert.Empty(t, profiles.Profiles[0].Stale)
	assert.Equal(t, 1, profiles.Profiles[0].Pinned)

	var sets polyformmcp.ListVariantSetsOutput
	callTool(t, session, "list_variant_sets", map[string]any{}, &sets)
	require.Len(t, sets.VariantSets, 1)
	assert.Equal(t, "Both", sets.VariantSets[0].Name)
	require.Len(t, sets.VariantSets[0].Dimensions, 1)
	assert.Equal(t, "Count", sets.VariantSets[0].Dimensions[0].Path)

	path := t.TempDir() + "/graph.json"
	callTool(t, session, "save_graph", map[string]any{"path": path}, &polyformmcp.SaveGraphOutput{})
	callTool(t, session, "load_graph", map[string]any{"path": path}, &polyformmcp.LoadGraphOutput{})
}

func TestUpdateVariableAcceptsAQuotedNumber(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": `"2.5"`},
			{"path": "Label", "type": "string", "value": `"5"`},
		},
	}, &polyformmcp.CreateVariablesOutput{})

	callTool(t, session, "update_variable", map[string]any{"path": "Width", "value": `"7"`}, &polyformmcp.UpdateVariableOutput{})

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	byPath := map[string]any{}
	for _, v := range list.Variables {
		byPath[v.Path] = v.Value
	}
	assert.Equal(t, 7.0, byPath["Width"])
	assert.Equal(t, "5", byPath["Label"], "a string variable keeps the string")

	msg := callToolExpectingError(t, session, "update_variable", map[string]any{"path": "Width", "value": `"red"`})
	assert.Contains(t, msg, "red")
}

func TestUpdateVariablesAppliesEachEntryAndReportsTheRest(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{
			{"path": "Width", "type": "float64", "value": "1"},
			{"path": "Count", "type": "int", "value": "1"},
		},
	}, &polyformmcp.CreateVariablesOutput{})

	var out polyformmcp.UpdateVariablesOutput
	callTool(t, session, "update_variables", map[string]any{"variables": []map[string]any{
		{"path": "Width", "value": "3"},
		{"path": "Nope", "value": "3"},
		{"path": "Count", "reset": true},
		{"path": "Count", "description": "how many"},
	}}, &out)
	require.Len(t, out.Results, 4)
	assert.True(t, out.Results[0].Updated)
	assert.Contains(t, out.Results[1].Error, "Nope")
	assert.Contains(t, out.Results[2].Error, "no recorded default")
	assert.True(t, out.Results[3].Updated)
	assert.Equal(t, 2, out.Failed)

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	for _, v := range list.Variables {
		switch v.Path {
		case "Width":
			assert.Equal(t, 3.0, v.Value)
		case "Count":
			assert.Equal(t, 1.0, v.Value)
			assert.Equal(t, "how many", v.Description)
		}
	}

	msg := callToolExpectingError(t, session, "update_variables", map[string]any{"variables": []map[string]any{{"path": "Nope", "value": "1"}}})
	assert.Contains(t, msg, "Nope")
}
