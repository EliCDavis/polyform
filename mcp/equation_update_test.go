package mcp_test

import (
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const intParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[int]"

func equationInstance(t *testing.T, session *mcpsdk.ClientSession, id, equation string, inputs map[string]float64) string {
	t.Helper()
	callTool(t, session, "create_equation_subgraph", map[string]any{"id": id, "equation": equation}, &polyformmcp.CreateEquationSubgraphOutput{})
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": id}, &placed)
	for name, value := range inputs {
		callTool(t, session, "connect_nodes", map[string]any{
			"outNodeId": literalFloat64Node(t, session, value), "outPort": "Value",
			"inNodeId": placed.NodeId, "inPort": name,
		}, &polyformmcp.ConnectNodesOutput{})
	}
	return placed.NodeId
}

func TestUpdateEquationSubgraphKeepsWiredInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = a + b", map[string]float64{"a": 3, "b": 4})
	require.InDelta(t, 7, evalFloat64Output(t, inst, placed, "y"), 1e-9)

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "equation": "y = a - b"}, &out)

	assert.Equal(t, []string{"a", "b"}, out.Inputs)
	assert.Empty(t, out.AddedInputs)
	assert.Empty(t, out.RemovedInputs)
	assert.Empty(t, out.Dropped)
	assert.InDelta(t, -1, evalFloat64Output(t, inst, placed, "y"), 1e-9)
}

func TestUpdateEquationSubgraphAddsAndRemovesInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = a + b", map[string]float64{"a": 3, "b": 4})

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "equation": "y = a * c"}, &out)

	assert.Equal(t, []string{"a", "c"}, out.Inputs)
	assert.Equal(t, []string{"c"}, out.AddedInputs)
	assert.Equal(t, []string{"b"}, out.RemovedInputs)
	assert.Equal(t, []string{"/" + placed}, out.Instances)
	require.Len(t, out.Dropped, 1, "the wire into the removed input is reported")
	assert.Equal(t, placed+".b", out.Dropped[0].To)

	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": literalFloat64Node(t, session, 5), "outPort": "Value",
		"inNodeId": placed, "inPort": "c",
	}, &polyformmcp.ConnectNodesOutput{})
	assert.InDelta(t, 15, evalFloat64Output(t, inst, placed, "y"), 1e-9)
}

func TestUpdateEquationSubgraphRenamesASingleOutputKeepingConsumers(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = a * 2", map[string]float64{"a": 3})

	var consumer polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   multiplyNodeType,
		"inputs": map[string]any{"Values": map[string]any{"elements": []map[string]any{{"nodeId": placed, "port": "y"}}}},
	}, &consumer)

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "equation": "z = a * 3"}, &out)

	assert.Equal(t, "y -> z", out.RenamedOutput)
	assert.Equal(t, "z", out.Output)
	assert.Empty(t, out.Dropped)
	assert.InDelta(t, 9, evalFloat64Output(t, inst, consumer.NodeId, "Float"), 1e-9)
}

func TestUpdateEquationSubgraphRejectsBadEquationAndLeavesTheOldOne(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = a + b", map[string]float64{"a": 3, "b": 4})

	callToolExpectingError(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "equation": "y = pow(a, b)"})

	assert.InDelta(t, 7, evalFloat64Output(t, inst, placed, "y"), 1e-9)
}

func TestCreateEquationSubgraphConstantsBecomeDefaultedInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id": "eq", "equation": "y = k * x", "constants": map[string]any{"k": 0.5},
	}, &eq)
	assert.Equal(t, []string{"k", "x"}, eq.Inputs)
	require.Contains(t, eq.Defaults, "k")

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "eq", "inputs": map[string]any{"x": map[string]any{"value": "4"}},
	}, &placed)
	assert.InDelta(t, 2, evalFloat64Output(t, inst, placed.NodeId, "y"), 1e-9)

	var updated polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "equation": "y = k * x", "constants": map[string]any{"k": 0.25},
	}, &updated)
	assert.Equal(t, eq.Defaults["k"], updated.Defaults["k"], "the default literal is updated, not replaced")
	assert.InDelta(t, 1, evalFloat64Output(t, inst, placed.NodeId, "y"), 1e-9)

	msg := callToolExpectingError(t, session, "create_equation_subgraph", map[string]any{
		"id": "bad", "equation": "y = x * 2", "constants": map[string]any{"k": 0.5},
	})
	assert.Contains(t, msg, "not an input")
}

func TestCreateEquationSubgraphMultipleStatementsShareInputs(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id": "polar", "equation": "r = sqrt(x^2 + y^2); d = r * 2",
	}, &eq)
	assert.Equal(t, []string{"x", "y"}, eq.Inputs)
	assert.Equal(t, "r", eq.Output)
	assert.Equal(t, []string{"r", "d"}, eq.Outputs)

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "polar",
		"inputs":     map[string]any{"x": map[string]any{"value": "3"}, "y": map[string]any{"value": "4"}},
	}, &placed)
	assert.InDelta(t, 5, evalFloat64Output(t, inst, placed.NodeId, "r"), 1e-9)
	assert.InDelta(t, 10, evalFloat64Output(t, inst, placed.NodeId, "d"), 1e-9)

	msg := callToolExpectingError(t, session, "create_equation_subgraph", map[string]any{
		"id": "twice", "equation": "a = x; a = y",
	})
	assert.Contains(t, msg, "defined twice")
}

func TestCreateEquationSubgraphIntPorts(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var eq polyformmcp.CreateEquationSubgraphOutput
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id": "eq", "equation": "n = i * x / 4", "types": map[string]any{"i": "int", "n": "int"},
	}, &eq)

	var count polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": intParamType}, &count)
	callTool(t, session, "set_parameter", map[string]any{"nodeId": count.NodeId, "value": "3"}, &polyformmcp.SetParameterOutput{})

	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "eq",
		"inputs": map[string]any{
			"i": map[string]any{"nodeId": count.NodeId, "port": "Value"},
			"x": map[string]any{"value": "3"},
		},
	}, &placed)

	out, ok := outputByLabel(t, inst.Node(placed.NodeId), "n").(nodes.Output[int])
	require.True(t, ok, "the output is typed int")
	assert.Equal(t, 2, out.Value(), "3*3/4 = 2.25 rounds to 2")

	msg := callToolExpectingError(t, session, "create_equation_subgraph", map[string]any{
		"id": "bad", "equation": "y = x", "types": map[string]any{"q": "int"},
	})
	assert.Contains(t, msg, "neither an input nor an output")
}

func TestUpdateEquationSubgraphPatchesOneStatement(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "s = a + b; d = a - b; p = a * b", map[string]float64{"a": 6, "b": 2})

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "statements": map[string]any{"d": "b - a"},
	}, &out)

	assert.Empty(t, out.Dropped)
	assert.InDelta(t, 8, evalFloat64Output(t, inst, placed, "s"), 1e-9)
	assert.InDelta(t, -4, evalFloat64Output(t, inst, placed, "d"), 1e-9)
	assert.InDelta(t, 12, evalFloat64Output(t, inst, placed, "p"), 1e-9)

	var read polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq"}, &read)
	assert.Equal(t, "s = a + b\nd = b - a\np = a * b", read.Equation)
}

func TestUpdateEquationSubgraphAddsStatementsBeforeTheirReaders(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = a + 1", map[string]float64{"a": 2})

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "statements": map[string]any{"y": "half + 1", "half": "a / 2"},
	}, &out)

	assert.Equal(t, []string{"a"}, out.Inputs, "half is an output read by y, not a new input")
	assert.InDelta(t, 2, evalFloat64Output(t, inst, placed, "y"), 1e-9)
	assert.InDelta(t, 1, evalFloat64Output(t, inst, placed, "half"), 1e-9)
}

func TestUpdateEquationSubgraphRemovesAStatement(t *testing.T) {
	session := testSession(t)
	equationInstance(t, session, "eq", "s = a + b; d = a - b", nil)

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "remove": []string{"d"}}, &out)
	assert.Equal(t, []string{"d"}, out.RemovedOutputs)

	msg := callToolExpectingError(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "remove": []string{"nope"}})
	assert.Contains(t, msg, `no statement defines "nope"`)
	msg = callToolExpectingError(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "equation": "s = a", "statements": map[string]any{"s": "b"},
	})
	assert.Contains(t, msg, "not both")
}

func TestUpdateEquationSubgraphKeepsIntPortsItWasNotTold(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id": "eq", "equation": "n = i * 2; x = i / 2", "types": map[string]any{"i": "int", "n": "int"},
	}, &polyformmcp.CreateEquationSubgraphOutput{})
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "eq", "inputs": map[string]any{"i": map[string]any{"value": "3"}},
	}, &placed)

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "statements": map[string]any{"x": "i / 4"}}, &out)

	assert.Empty(t, out.Dropped, "the int input kept its type and its wire")
	assert.Equal(t, "int", inputByLabel(t, inst.Node(placed.NodeId), "i").(nodes.Typed).Type())
	assert.InDelta(t, 0.75, evalFloat64Output(t, inst, placed.NodeId, "x"), 1e-9)
}

func TestEquationStatementsAreOrderedByWhatTheyRead(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = half + 1; half = a / 2", map[string]float64{"a": 4})

	assert.InDelta(t, 3, evalFloat64Output(t, inst, placed, "y"), 1e-9)

	msg := callToolExpectingError(t, session, "create_equation_subgraph", map[string]any{"id": "loop", "equation": "a = b + 1; b = a + 1"})
	assert.Contains(t, msg, "cycle")
}

func TestUpdateEquationSubgraphKeepsLiftedConsumersDownstream(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_equation_subgraph", map[string]any{"id": "eq", "equation": "y = a * 2"}, &polyformmcp.CreateEquationSubgraphOutput{})
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "array", "type": floatArrayParamType, "value": "[1, 2, 3]"},
		map[string]any{"alias": "eq", "subgraphId": "eq", "inputs": map[string]any{
			"a": map[string]any{"nodeId": "array", "port": "Value"},
		}},
		map[string]any{"alias": "round", "type": roundNodeType, "inputs": map[string]any{
			"In": map[string]any{"nodeId": "eq", "port": "y"},
		}},
		map[string]any{"alias": "lines", "type": varyingRadiusLinesType, "inputs": map[string]any{
			"Radii": map[string]any{"nodeId": "round", "port": "Float"},
		}},
	}}, &made)
	require.Empty(t, made.Errors)

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{"id": "eq", "statements": map[string]any{"y": "a * 3"}}, &out)
	assert.Empty(t, out.Dropped, "rebuilding the interior must not drop wires two nodes downstream")
}

func TestStoredEquationSurvivesSaveAndLoad(t *testing.T) {
	session := testSession(t)
	equationInstance(t, session, "eq", "s = a + b; d = a - b", nil)
	path := filepath.Join(t.TempDir(), "graph.json")
	callTool(t, session, "save_graph", map[string]any{"path": path}, nil)

	reloaded := testSession(t)
	callTool(t, reloaded, "load_graph", map[string]any{"path": path}, nil)
	var read polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, reloaded, "update_equation_subgraph", map[string]any{"id": "eq"}, &read)
	assert.Equal(t, "s = a + b\nd = a - b", read.Equation)
}

func TestUpdateEquationSubgraphRecreatesAPortWhoseTypeChanges(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	placed := equationInstance(t, session, "eq", "y = i * 2", map[string]float64{"i": 3})

	var out polyformmcp.UpdateEquationSubgraphOutput
	callTool(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "equation": "y = i * 2", "types": map[string]any{"i": "int"},
	}, &out)

	assert.Equal(t, []string{"i"}, out.AddedInputs)
	assert.Equal(t, []string{"i"}, out.RemovedInputs)
	require.Len(t, out.Dropped, 1, "the float literal no longer fits the int port")
	assert.Equal(t, "int", inputByLabel(t, inst.Node(placed), "i").(nodes.Typed).Type())
	assert.InDelta(t, 0, evalFloat64Output(t, inst, placed, "y"), 1e-9)
}
