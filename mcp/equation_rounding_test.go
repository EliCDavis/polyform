package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Indexing a copy into a repeating pattern - which board of a wall, which
// tooth of a gear - needs floor and mod, and a count-driven graph has no
// other way to get them.
func TestCreateEquationSubgraphRoundingAndModulo(t *testing.T) {
	cases := map[string]struct {
		equation string
		inputs   map[string]float64
		expected float64
	}{
		"floor":              {"y = floor(x)", map[string]float64{"x": 2.7}, 2},
		"floor negative":     {"y = floor(x)", map[string]float64{"x": -2.1}, -3},
		"ceil":               {"y = ceil(x)", map[string]float64{"x": 2.1}, 3},
		"round":              {"y = round(x)", map[string]float64{"x": 2.5}, 3},
		"abs":                {"y = abs(x)", map[string]float64{"x": -4.5}, 4.5},
		"mod":                {"y = mod(a, b)", map[string]float64{"a": 7, "b": 3}, 1},
		"mod stays positive": {"y = mod(a, b)", map[string]float64{"a": -1, "b": 3}, 2},
		"every other board":  {"y = mod(floor(i), 2)", map[string]float64{"i": 5.4}, 1},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			session, inst := testSessionWithInstance(t)

			var eq polyformmcp.CreateEquationSubgraphOutput
			callTool(t, session, "create_equation_subgraph", map[string]any{
				"id": "eq", "equation": c.equation,
			}, &eq)

			var placed polyformmcp.InstantiateSubgraphOutput
			callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "eq"}, &placed)

			for variable, value := range c.inputs {
				id := literalFloat64Node(t, session, value)
				callTool(t, session, "connect_nodes", map[string]any{
					"outNodeId": id, "outPort": "Value",
					"inNodeId": placed.NodeId, "inPort": variable,
				}, &polyformmcp.ConnectNodesOutput{})
			}

			assert.InDelta(t, c.expected, evalFloat64Output(t, inst, placed.NodeId, "y"), 1e-9)
		})
	}
}

func TestCreateEquationSubgraphStillNamesUnsupportedFunctions(t *testing.T) {
	session := testSession(t)
	msg := callToolExpectingError(t, session, "create_equation_subgraph", map[string]any{
		"id": "bad", "equation": "y = pow(x, 3)",
	})
	require.Contains(t, msg, "pow")
	assert.Contains(t, msg, "floor(x)", "the error lists what is supported")
}
