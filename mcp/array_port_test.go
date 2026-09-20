package mcp_test

import (
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const multiplyNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.MultiplyNode[float64]]"

func multiplyOf(t *testing.T, session *mcpsdk.ClientSession, values ...string) string {
	t.Helper()
	elements := make([]map[string]any, 0, len(values))
	for _, v := range values {
		elements = append(elements, map[string]any{"value": v})
	}
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   multiplyNodeType,
		"inputs": map[string]any{"Values": map[string]any{"elements": elements}},
	}, &created)
	return created.NodeId
}

func TestConnectNodesIndexedReplacesElementInPlace(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	mul := multiplyOf(t, session, "2", "3", "5")
	seven := literalFloat64Node(t, session, 7)

	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": seven, "outPort": "Value",
		"inNodeId": mul, "inPort": "Values.1",
	}, &polyformmcp.ConnectNodesOutput{})

	assert.InDelta(t, 2*7*5, evalFloat64Output(t, inst, mul, "Float"), 1e-9)
}

func TestConnectNodesIndexAtEndAppends(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	mul := multiplyOf(t, session, "2", "3")
	seven := literalFloat64Node(t, session, 7)

	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": seven, "outPort": "Value",
		"inNodeId": mul, "inPort": "Values.2",
	}, &polyformmcp.ConnectNodesOutput{})

	assert.InDelta(t, 2*3*7, evalFloat64Output(t, inst, mul, "Float"), 1e-9)
}

func TestConnectNodesIndexBeyondEndIsError(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	mul := multiplyOf(t, session, "2", "3")
	seven := literalFloat64Node(t, session, 7)

	msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
		"outNodeId": seven, "outPort": "Value",
		"inNodeId": mul, "inPort": "Values.5",
	})
	assert.Contains(t, msg, "2 element(s)")
	assert.InDelta(t, 6, evalFloat64Output(t, inst, mul, "Float"), 1e-9, "a refused connect must leave the port untouched")
}

func TestConnectNodesIndexOnSingleValuePortIsError(t *testing.T) {
	session := testSession(t)
	var sub polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": subtractNodeType}, &sub)
	seven := literalFloat64Node(t, session, 7)

	msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
		"outNodeId": seven, "outPort": "Value",
		"inNodeId": sub.NodeId, "inPort": "A.0",
	})
	assert.Contains(t, msg, "single value")
}

func TestDisconnectReportsRemainingElements(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	mul := multiplyOf(t, session, "2", "3", "5")

	var out polyformmcp.DisconnectOutput
	callTool(t, session, "disconnect", map[string]any{"nodeId": mul, "port": "Values.0"}, &out)

	require.True(t, out.Disconnected)
	require.Len(t, out.Remaining, 2)
	for i, ref := range out.Remaining {
		assert.Equal(t, "Value", ref.Port)
		assert.NotEmpty(t, ref.NodeId, "remaining[%d] should name the node still wired in", i)
	}
	assert.InDelta(t, 15, evalFloat64Output(t, inst, mul, "Float"), 1e-9)
}

func TestDisconnectIndexBeyondEndIsError(t *testing.T) {
	session := testSession(t)
	mul := multiplyOf(t, session, "2", "3")

	msg := callToolExpectingError(t, session, "disconnect", map[string]any{"nodeId": mul, "port": "Values.4"})
	assert.Contains(t, msg, "no index 4")
}

func TestCreateNodeRejectsIndexedInputKey(t *testing.T) {
	session := testSession(t)
	two := literalFloat64Node(t, session, 2)

	msg := callToolExpectingError(t, session, "create_node", map[string]any{
		"type": multiplyNodeType,
		"inputs": map[string]any{
			"Values.0": map[string]any{"nodeId": two, "port": "Value"},
		},
	})
	assert.Contains(t, msg, "elements")
}

func TestCreateNodesRejectsIndexedInputKey(t *testing.T) {
	session := testSession(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []map[string]any{
			{"alias": "two", "type": floatParamType},
			{"alias": "mul", "type": multiplyNodeType, "inputs": map[string]any{
				"Values.0": map[string]any{"nodeId": "two", "port": "Value"},
			}},
		},
	}, &out)

	require.Len(t, out.Errors, 1)
	assert.Contains(t, out.Errors[0], "elements")
}

func TestConnectNodesFromVariable(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	mul := multiplyOf(t, session, "3")
	callTool(t, session, "create_variable", map[string]any{
		"path": "Scale", "type": "float64", "value": "4",
	}, &polyformmcp.CreateVariableOutput{})

	callTool(t, session, "connect_nodes", map[string]any{
		"variable": "Scale",
		"inNodeId": mul, "inPort": "Values",
	}, &polyformmcp.ConnectNodesOutput{})
	assert.InDelta(t, 12, evalFloat64Output(t, inst, mul, "Float"), 1e-9)

	callTool(t, session, "update_variable", map[string]any{"path": "Scale", "value": "10"}, nil)
	assert.InDelta(t, 30, evalFloat64Output(t, inst, mul, "Float"), 1e-9, "the reference should follow the variable")
}

func TestConnectNodesVariableAndNodeAreExclusive(t *testing.T) {
	session := testSession(t)
	mul := multiplyOf(t, session, "3")
	two := literalFloat64Node(t, session, 2)

	msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
		"variable":  "Scale",
		"outNodeId": two, "outPort": "Value",
		"inNodeId": mul, "inPort": "Values",
	})
	assert.Contains(t, msg, "not both")
}

func TestDescribeGraphFiltersToNodesAndTheirInputs(t *testing.T) {
	session := testSession(t)
	mul := multiplyOf(t, session, "2", "3")
	unrelated := multiplyOf(t, session, "5")

	var out polyformmcp.DescribeGraphOutput
	callTool(t, session, "describe_graph", map[string]any{"nodeIds": []string{mul}}, &out)

	ids := map[string]bool{}
	for _, n := range out.Nodes {
		ids[n.Id] = true
	}
	assert.True(t, ids[mul])
	assert.False(t, ids[unrelated])
	assert.Len(t, out.Nodes, 3, "the node plus the two literals feeding it")
}

func TestDescribeGraphFilterUnknownIdIsError(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "describe_graph", map[string]any{"nodeIds": []string{"Node-999"}})
	assert.Contains(t, msg, "Node-999")
}

func TestGenerateDefaultsToProjectDist(t *testing.T) {
	session := testSession(t)
	dir := t.TempDir()
	callTool(t, session, "start_project", map[string]any{"path": dir}, &polyformmcp.StartProjectOutput{})

	var text polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   textNodeType,
		"inputs": map[string]any{"In": map[string]any{"value": `"hello"`}},
	}, &text)
	callTool(t, session, "set_producer", map[string]any{
		"nodeId": text.NodeId, "port": "Out", "name": "hello.txt",
	}, nil)

	var out polyformmcp.GenerateOutput
	callTool(t, session, "generate", map[string]any{}, &out)

	assert.Equal(t, filepath.Join(dir, "dist"), out.OutputDir)
	require.NotEmpty(t, out.Files)
	for _, f := range out.Files {
		_, err := os.Stat(f)
		assert.NoError(t, err)
	}
}

func TestGenerateWithoutProjectRequiresOutputDir(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "generate", map[string]any{})
	assert.Contains(t, msg, "outputDir")
}
