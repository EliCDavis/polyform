package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Result = 0.5 * (whatever is appended to the multiply's Values later).
func halfSubgraph(t *testing.T, session *mcpsdk.ClientSession, id string) polyformmcp.CreateSubgraphOutput {
	t.Helper()
	var out polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id":   id,
		"name": "Half",
		"nodes": []map[string]any{
			{"alias": "m", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{{"value": "0.5"}}},
			}},
		},
		"outputs": []map[string]any{{"name": "Result", "type": "float64", "nodeId": "m", "port": "Float"}},
	}, &out)
	return out
}

func TestCreateBoundaryNodeWiresEveryExistingInstance(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	half := halfSubgraph(t, session, "half")
	var first, second polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "half"}, &first)
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "half"}, &second)
	callTool(t, session, "create_variables", map[string]any{"variables": []map[string]any{
		{"path": "Size", "type": "float64", "value": "4"},
	}}, &polyformmcp.CreateVariablesOutput{})

	var out polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half", "kind": "input", "portType": "float64", "name": "X",
		"wireInstances": map[string]any{"variable": "Size"},
	}, &out)
	assert.ElementsMatch(t, []string{"/" + first.NodeId, "/" + second.NodeId}, out.Wired)
	assert.Empty(t, out.Unwired)

	callTool(t, session, "connect_nodes", map[string]any{
		"scope": "half", "outNodeId": out.NodeId, "outPort": "Value", "inNodeId": half.Nodes["m"], "inPort": "Values",
	}, &polyformmcp.ConnectNodesOutput{})

	assert.InDelta(t, 2, evalFloat64Output(t, inst, first.NodeId, "Result"), 1e-9)
	assert.InDelta(t, 2, evalFloat64Output(t, inst, second.NodeId, "Result"), 1e-9)
}

func TestCreateBoundaryNodeBatchWithLiteralDefaults(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	half := halfSubgraph(t, session, "half")
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "half"}, &placed)

	var out polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half",
		"ports": []map[string]any{
			{"kind": "input", "portType": "float64", "name": "X", "wireInstances": map[string]any{"value": "4"}},
			{"kind": "input", "portType": "float64", "name": "Y", "wireInstances": map[string]any{"value": "3"}},
			{"kind": "output", "portType": "float64", "name": "Twice"},
		},
	}, &out)
	require.Len(t, out.Ports, 3)
	byName := map[string]polyformmcp.CreatedBoundaryPort{}
	for _, port := range out.Ports {
		byName[port.Name] = port
	}
	assert.Equal(t, []string{"/" + placed.NodeId}, byName["X"].Wired)
	assert.Equal(t, []string{"/" + placed.NodeId}, byName["Y"].Wired)
	assert.Empty(t, byName["Twice"].Instances)

	callTool(t, session, "connect_nodes", map[string]any{"scope": "half", "connections": []map[string]any{
		{"outNodeId": byName["X"].NodeId, "outPort": "Value", "inNodeId": half.Nodes["m"], "inPort": "Values"},
		{"outNodeId": byName["Y"].NodeId, "outPort": "Value", "inNodeId": half.Nodes["m"], "inPort": "Values"},
	}}, &polyformmcp.ConnectNodesOutput{})
	assert.InDelta(t, 6, evalFloat64Output(t, inst, placed.NodeId, "Result"), 1e-9)

	msg := callToolExpectingError(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half",
		"ports": []map[string]any{
			{"kind": "input", "portType": "float64", "name": "Z"},
			{"kind": "sideways", "portType": "float64", "name": "W"},
		},
	})
	assert.Contains(t, msg, "ports[1]")
	var listed polyformmcp.ListNodesOutput
	callTool(t, session, "list_nodes", map[string]any{"scope": "half"}, &listed)
	for _, node := range listed.Nodes {
		assert.NotEqual(t, "Z", node.Name, "a failed batch creates nothing")
	}
}

func TestCreateSubgraphInputDefaultIsReadByAnUnwiredInstance(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "scaled", "name": "Scaled",
		"inputs": []map[string]any{{"name": "Scale", "type": "float64", "default": "4"}},
		"nodes": []map[string]any{
			{"alias": "m", "type": multiplyNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{{"value": "3"}, {"nodeId": "Scale", "port": "Value"}}},
			}},
		},
		"outputs": []map[string]any{{"name": "Result", "type": "float64", "nodeId": "m", "port": "Float"}},
	}, &polyformmcp.CreateSubgraphOutput{})

	var unwired, wired polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "scaled"}, &unwired)
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled", "inputs": map[string]any{"Scale": map[string]any{"value": "10"}},
	}, &wired)

	assert.InDelta(t, 12, evalFloat64Output(t, inst, unwired.NodeId, "Result"), 1e-9)
	assert.InDelta(t, 30, evalFloat64Output(t, inst, wired.NodeId, "Result"), 1e-9)
}

func TestCreateBoundaryNodeDefaultsToTheVariableItWires(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	half := halfSubgraph(t, session, "half")
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "half"}, &placed)
	callTool(t, session, "create_variables", map[string]any{"variables": []map[string]any{
		{"path": "Size", "type": "float64", "value": "4"},
	}}, &polyformmcp.CreateVariablesOutput{})

	var out polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half", "kind": "input", "portType": "float64", "name": "Size",
		"wireInstances": map[string]any{"variable": "Size"},
	}, &out)
	callTool(t, session, "connect_nodes", map[string]any{
		"scope": "half", "outNodeId": out.NodeId, "outPort": "Value", "inNodeId": half.Nodes["m"], "inPort": "Values",
	}, &polyformmcp.ConnectNodesOutput{})

	var unwired polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "half"}, &unwired)
	assert.InDelta(t, 2, evalFloat64Output(t, inst, unwired.NodeId, "Result"), 1e-9)

	msg := callToolExpectingError(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half", "kind": "output", "portType": "float64", "name": "Other", "default": "1",
	})
	assert.Contains(t, msg, "only apply to an input")
}

func TestCreateBoundaryNodeThroughNestedSubgraphs(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	half := halfSubgraph(t, session, "half")
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "mid", "name": "Mid",
		"nodes":   []map[string]any{{"alias": "h", "subgraphId": "half"}},
		"outputs": []map[string]any{{"name": "Result", "type": "float64", "nodeId": "h", "port": "Result"}},
	}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "outer", "name": "Outer",
		"nodes":   []map[string]any{{"alias": "m", "subgraphId": "mid"}},
		"outputs": []map[string]any{{"name": "Result", "type": "float64", "nodeId": "m", "port": "Result"}},
	}, &polyformmcp.CreateSubgraphOutput{})
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "outer"}, &placed)

	var out polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half", "kind": "input", "portType": "float64", "name": "X",
		"through":       []string{"outer", "mid"},
		"wireInstances": map[string]any{"value": "4"},
	}, &out)
	assert.Len(t, out.Chain, 2)
	assert.Contains(t, out.Chain, "outer")
	assert.Contains(t, out.Chain, "mid")
	assert.Equal(t, []string{"/" + placed.NodeId}, out.Wired)

	callTool(t, session, "connect_nodes", map[string]any{
		"scope": "half", "outNodeId": out.NodeId, "outPort": "Value", "inNodeId": half.Nodes["m"], "inPort": "Values",
	}, &polyformmcp.ConnectNodesOutput{})
	assert.InDelta(t, 2, evalFloat64Output(t, inst, placed.NodeId, "Result"), 1e-9)

	var subgraphs polyformmcp.ListSubgraphsOutput
	callTool(t, session, "list_subgraphs", map[string]any{}, &subgraphs)
	for _, sg := range subgraphs.Subgraphs {
		if sg.Id == "half" {
			require.Len(t, sg.Instances, 1)
			assert.Regexp(t, "^"+placed.NodeId+"/Node-[0-9]+/Node-[0-9]+$", sg.Instances[0])
			require.Len(t, sg.Placements, 1)
			assert.Regexp(t, "^mid/Node-[0-9]+$", sg.Placements[0])
		}
	}

	msg := callToolExpectingError(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half", "kind": "input", "portType": "float64", "name": "Y",
		"through": []string{"outer"},
	})
	assert.Contains(t, msg, "places no instance")
}

func TestCreateBoundaryNodeThroughCarriesTheInputsOfABatchThatAlsoAddsOutputs(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	halfSubgraph(t, session, "half")
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "outer", "name": "Outer",
		"nodes": []map[string]any{{"alias": "h", "subgraphId": "half"}},
	}, &polyformmcp.CreateSubgraphOutput{})
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "outer"}, &polyformmcp.InstantiateSubgraphOutput{})

	var out polyformmcp.CreateBoundaryNodeOutput
	callTool(t, session, "create_boundary_node", map[string]any{
		"subgraphId": "half",
		"through":    []string{"outer"},
		"ports": []map[string]any{
			{"kind": "input", "portType": "float64", "name": "X"},
			{"kind": "output", "portType": "float64", "name": "Echo"},
		},
	}, &out)
	require.Len(t, out.Ports, 2)
	assert.Contains(t, out.Ports[0].Chain, "outer")
	assert.Empty(t, out.Ports[1].Chain)
}

func TestCreateBoundaryNodeRefusesSinglePortFieldsBesideAPortsList(t *testing.T) {
	session := testSession(t)
	halfSubgraph(t, session, "half")

	msg := callToolExpectingError(t, session, "create_boundary_node", map[string]any{
		"subgraphId":    "half",
		"ports":         []map[string]any{{"kind": "input", "name": "X", "portType": "float64"}},
		"wireInstances": map[string]any{"value": "4"},
	})

	assert.Contains(t, msg, `"wireInstances" belongs inside each "ports" entry`)
}
