package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const negateNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.NegateNode[float64]]"

// Result = Base + Offset.
func offsetSubgraph(t *testing.T, session *mcpsdk.ClientSession, id string) {
	t.Helper()
	const addNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.AddNode[float64]]"
	callTool(t, session, "create_subgraph", map[string]any{
		"id":   id,
		"name": "Offset",
		"inputs": []map[string]any{
			{"name": "Base", "type": "float64"},
			{"name": "Offset", "type": "float64"},
		},
		"nodes": []map[string]any{
			{"alias": "sum", "type": addNodeType, "inputs": map[string]any{
				"Values": map[string]any{"elements": []map[string]any{
					{"nodeId": "Base", "port": "Value"},
					{"nodeId": "Offset", "port": "Value"},
				}},
			}},
		},
		"outputs": []map[string]any{
			{"name": "Result", "type": "float64", "nodeId": "sum", "port": "Float"},
		},
	}, &polyformmcp.CreateSubgraphOutput{})
}

func TestReplaceInstanceKeepsMatchingWiringUnderTheSameId(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	scaledSubgraph(t, session, "scaled")
	offsetSubgraph(t, session, "offset")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled",
		"inputs": map[string]any{
			"Base":  map[string]any{"value": "3"},
			"Scale": map[string]any{"value": "4"},
		},
	}, &placed)
	var consumer polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   negateNodeType,
		"inputs": map[string]any{"In": map[string]any{"nodeId": placed.NodeId, "port": "Result"}},
	}, &consumer)
	require.InDelta(t, -12, evalFloat64Output(t, inst, consumer.NodeId, "Out"), 1e-9)

	var out polyformmcp.ReplaceInstanceOutput
	callTool(t, session, "replace_instance", map[string]any{"nodeId": placed.NodeId, "subgraphId": "offset"}, &out)

	assert.Equal(t, placed.NodeId, out.NodeId)
	assert.Equal(t, []string{"Base"}, out.KeptInputs)
	assert.Equal(t, []string{"Result -> " + consumer.NodeId + ".In"}, out.KeptOutputs)
	require.Len(t, out.LostInputs, 1)
	assert.Contains(t, out.LostInputs[0], "Scale")
	assert.Equal(t, []string{"Offset"}, out.OpenInputs)
	assert.InDelta(t, -3, evalFloat64Output(t, inst, consumer.NodeId, "Out"), 1e-9, "Base carried over, Offset reads zero")

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{placed.NodeId}, "nodeOnly": true}, &desc)
	require.Len(t, desc.Nodes, 1)
	assert.Equal(t, "offset", desc.Nodes[0].SubGraphId)
}

func TestReplaceInstanceKeepsArrayElementsInPlace(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	scaledSubgraph(t, session, "scaled")
	offsetSubgraph(t, session, "offset")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled",
		"inputs":     map[string]any{"Base": map[string]any{"value": "2"}, "Scale": map[string]any{"value": "5"}},
	}, &placed)
	var product polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": multiplyNodeType,
		"inputs": map[string]any{"Values": map[string]any{"elements": []map[string]any{
			{"value": "100"},
			{"nodeId": placed.NodeId, "port": "Result"},
			{"value": "7"},
		}}},
	}, &product)
	require.InDelta(t, 100*10*7, evalFloat64Output(t, inst, product.NodeId, "Float"), 1e-9)

	callTool(t, session, "replace_instance", map[string]any{"nodeId": placed.NodeId, "subgraphId": "offset"}, &polyformmcp.ReplaceInstanceOutput{})

	var desc polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{product.NodeId}, "nodeOnly": true}, &desc)
	assert.Equal(t, placed.NodeId, desc.Nodes[0].AssignedInput["Values.1"].NodeId, "the swapped instance holds its slot")
	assert.InDelta(t, 100*2*7, evalFloat64Output(t, inst, product.NodeId, "Float"), 1e-9)
}

func TestReplaceInstanceRefusesANonInstance(t *testing.T) {
	session := testSession(t)
	scaledSubgraph(t, session, "scaled")
	literal := literalFloat64Node(t, session, 1)
	msg := callToolExpectingError(t, session, "replace_instance", map[string]any{"nodeId": literal, "subgraphId": "scaled"})
	assert.Contains(t, msg, "not a subgraph instance")
}

func TestInstantiateSubgraphCopiesInputsAcrossSubgraphsByPortName(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	scaledSubgraph(t, session, "scaled")
	offsetSubgraph(t, session, "offset")

	var original polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "scaled",
		"inputs":     map[string]any{"Base": map[string]any{"value": "3"}, "Scale": map[string]any{"value": "4"}},
	}, &original)

	var copied polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId":     "offset",
		"copyInputsFrom": original.NodeId,
		"inputs":         map[string]any{"Offset": map[string]any{"value": "10"}},
	}, &copied)

	assert.Equal(t, []string{"Base"}, copied.Copied)
	require.Len(t, copied.Skipped, 1)
	assert.Contains(t, copied.Skipped[0], "Scale")
	assert.InDelta(t, 13, evalFloat64Output(t, inst, copied.NodeId, "Result"), 1e-9)
}
