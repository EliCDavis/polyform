package mcp_test

import (
	"strings"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func equationNodeCounts(t *testing.T, session *mcpsdk.ClientSession, id string) map[string]int {
	t.Helper()
	var listed polyformmcp.ListNodesOutput
	callTool(t, session, "list_nodes", map[string]any{"scope": id}, &listed)
	counts := map[string]int{}
	for _, n := range listed.Nodes {
		counts[n.Type]++
	}
	return counts
}

func countOfType(counts map[string]int, fragment string) int {
	total := 0
	for typ, n := range counts {
		if strings.Contains(typ, fragment) {
			total += n
		}
	}
	return total
}

func placeEquation(t *testing.T, session *mcpsdk.ClientSession, equation string, types map[string]string, inputs map[string]float64) string {
	t.Helper()
	args := map[string]any{"id": "eq", "equation": equation}
	if types != nil {
		args["types"] = types
	}
	callTool(t, session, "create_equation_subgraph", args, &polyformmcp.CreateEquationSubgraphOutput{})
	var placed polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "eq"}, &placed)
	for variable, value := range inputs {
		id := literalFloat64Node(t, session, value)
		callTool(t, session, "connect_nodes", map[string]any{
			"outNodeId": id, "outPort": "Value",
			"inNodeId": placed.NodeId, "inPort": variable,
		}, &polyformmcp.ConnectNodesOutput{})
	}
	return placed.NodeId
}

func TestEquationBuildsARepeatedSubexpressionOnce(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id":       "eq",
		"equation": "lo = x/2 - w; hi = x/2 + w; mid = (x/2 - w + x/2 + w)/2; also = x/2 - w",
	}, &polyformmcp.CreateEquationSubgraphOutput{})

	counts := equationNodeCounts(t, session, "eq")
	assert.Equal(t, 2, countOfType(counts, "DivideNode"), "x/2 once, and the final /2")
	assert.Equal(t, 1, countOfType(counts, "SubtractNode"))
}

func TestEquationSharesAProductWrittenInEitherOrder(t *testing.T) {
	session := testSession(t)
	callTool(t, session, "create_equation_subgraph", map[string]any{
		"id": "eq", "equation": "y = a*b + b*a + min(a, b) + min(b, a)",
	}, &polyformmcp.CreateEquationSubgraphOutput{})

	counts := equationNodeCounts(t, session, "eq")
	assert.Equal(t, 1, countOfType(counts, "MultiplyNode"))
	assert.Equal(t, 1, countOfType(counts, "MinNode"))
}

func TestEquationKeepsOperandOrderWhereItMatters(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	id := placeEquation(t, session, "y = (a - b)*10 + (b - a)", nil, map[string]float64{"a": 5, "b": 2})
	assert.InDelta(t, 27, evalFloat64Output(t, inst, id, "y"), 1e-9)
}

func TestEquationFloorAndCeilShareOneRoundNode(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	id := placeEquation(t, session, "y = floor(x) + ceil(x)", nil, map[string]float64{"x": 2.5})
	assert.InDelta(t, 5, evalFloat64Output(t, inst, id, "y"), 1e-9)
	assert.Equal(t, 1, countOfType(equationNodeCounts(t, session, "eq"), "RoundNode"))
}

func TestEquationComparisonsAndIf(t *testing.T) {
	cases := map[string]struct {
		equation string
		inputs   map[string]float64
		expected float64
	}{
		"less, true":         {"y = if(x < 2, 10, 20)", map[string]float64{"x": 1}, 10},
		"less, false":        {"y = if(x < 2, 10, 20)", map[string]float64{"x": 3}, 20},
		"less at the edge":   {"y = if(x < 2, 10, 20)", map[string]float64{"x": 2}, 20},
		"less or equal":      {"y = if(x <= 2, 10, 20)", map[string]float64{"x": 2}, 10},
		"greater":            {"y = if(a > b, a, b)", map[string]float64{"a": 4, "b": 7}, 7},
		"greater or equal":   {"y = if(a >= b, a, b)", map[string]float64{"a": 7, "b": 7}, 7},
		"equal":              {"y = if(x == 2, 1, 0)", map[string]float64{"x": 2}, 1},
		"not equal":          {"y = if(x != 2, 1, 0)", map[string]float64{"x": 2}, 0},
		"and":                {"y = if(and(x > 0, x < 1), 1, 0)", map[string]float64{"x": 0.5}, 1},
		"and fails":          {"y = if(and(x > 0, x < 1), 1, 0)", map[string]float64{"x": 1.5}, 0},
		"or":                 {"y = if(or(x < 0, x > 1), 1, 0)", map[string]float64{"x": 1.5}, 1},
		"not":                {"y = if(not(x > 1), 1, 0)", map[string]float64{"x": 1.5}, 0},
		"tiny positive":      {"y = if(x > 0, 1, 0)", map[string]float64{"x": 1e-9}, 1},
		"arithmetic in both": {"y = if(a - b > 0, a - b, 0) * 2", map[string]float64{"a": 5, "b": 2}, 6},
		"nested if":          {"y = if(x < 0, -1, if(x > 0, 1, 0))", map[string]float64{"x": 0}, 0},
		"condition reused":   {"c = x > 2; y = if(c, 10, 0) + if(c, 1, 0)", map[string]float64{"x": 3}, 11},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			session, inst := testSessionWithInstance(t)
			id := placeEquation(t, session, c.equation, nil, c.inputs)
			assert.InDelta(t, c.expected, evalFloat64Output(t, inst, id, "y"), 1e-9)
		})
	}
}

func TestEquationIfLiftsOverAListCondition(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	id := placeEquation(t, session, "y = if(x == r, 1, 0)", nil, map[string]float64{"r": 0})

	var list polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": float64ArrayParamType}, &list)
	callTool(t, session, "set_parameter", map[string]any{"nodeId": list.NodeId, "value": "[0, 1]"}, &polyformmcp.SetParameterOutput{})
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": list.NodeId, "outPort": "Value", "inNodeId": id, "inPort": "x",
	}, &polyformmcp.ConnectNodesOutput{})

	out, ok := inst.Node(id).Outputs()["y"]
	if !ok {
		out = outputByLabel(t, inst.Node(id), "y")
	}
	valued, ok := out.(nodes.Output[[]float64])
	require.True(t, ok, "y is a list output, got %T", out)
	assert.Equal(t, []float64{1, 0}, valued.Value())
}

func TestEquationConditionIsABoolOutput(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	id := placeEquation(t, session, "show = and(w > 0.5, w < 3)", nil, map[string]float64{"w": 1})

	out, ok := inst.Node(id).Outputs()["show"]
	if !ok {
		out = outputByLabel(t, inst.Node(id), "show")
	}
	valued, ok := out.(nodes.Output[bool])
	require.True(t, ok, "show is a bool output")
	assert.True(t, valued.Value())
}

func TestEquationBoolInputDrivesIf(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	id := placeEquation(t, session, "y = if(on, a, 0)", map[string]string{"on": "bool"}, map[string]float64{"a": 4})

	var flag polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": "github.com/EliCDavis/polyform/generator/parameter.Value[bool]"}, &flag)
	callTool(t, session, "set_parameter", map[string]any{"nodeId": flag.NodeId, "value": "true"}, &polyformmcp.SetParameterOutput{})
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": flag.NodeId, "outPort": "Value", "inNodeId": id, "inPort": "on",
	}, &polyformmcp.ConnectNodesOutput{})

	assert.InDelta(t, 4, evalFloat64Output(t, inst, id, "y"), 1e-9)
}

func TestEquationRefusesMixingConditionsAndNumbers(t *testing.T) {
	cases := map[string]struct {
		equation string
		says     string
	}{
		"condition in arithmetic": {"y = (x > 1) * 3", "if(cond, 1, 0)"},
		"number as condition":     {"y = if(x, 1, 0)", "conditions"},
		"branches disagree":       {"y = if(x > 1, 1, x > 2)", "both"},
		"chained comparison":      {"y = if(0 < x < 1, 1, 0)", "and("},
		"single equals":           {"y = if(x = 1, 1, 0)", "=="},
		"not a number":            {"y = sqrt(x > 1)", "condition"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			session := testSession(t)
			msg := callToolExpectingError(t, session, "create_equation_subgraph", map[string]any{
				"id": "eq", "equation": c.equation,
			})
			assert.Contains(t, msg, c.says)
		})
	}
}

func TestUpdateEquationPatchesAStatementHoldingAComparison(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	id := placeEquation(t, session, "y = x*2", nil, map[string]float64{"x": 3})

	callTool(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "statements": map[string]string{"y": "if(x <= 2, 0, x*2)"},
	}, &polyformmcp.UpdateEquationSubgraphOutput{})
	assert.InDelta(t, 6, evalFloat64Output(t, inst, id, "y"), 1e-9)

	callTool(t, session, "update_equation_subgraph", map[string]any{
		"id": "eq", "statements": map[string]string{"y": "if(x == 3, 1, 0)"},
	}, &polyformmcp.UpdateEquationSubgraphOutput{})
	assert.InDelta(t, 1, evalFloat64Output(t, inst, id, "y"), 1e-9)
}
