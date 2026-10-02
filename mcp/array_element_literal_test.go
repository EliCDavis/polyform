package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const floatArrayParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[[]float64]"

// A lifted list port reports the array type as soon as an array reaches
// any element. A literal for a later element is still one value, so the
// order the elements are given in must not decide whether it parses.
func TestAListPortTakesANumberBesideAnArrayInEitherOrder(t *testing.T) {
	for _, arrayFirst := range []bool{false, true} {
		name := "number then array"
		if arrayFirst {
			name = "array then number"
		}
		t.Run(name, func(t *testing.T) {
			session := testSession(t)

			var arr polyformmcp.CreateNodeOutput
			callTool(t, session, "create_node", map[string]any{"type": floatArrayParamType}, &arr)
			var set polyformmcp.SetParameterOutput
			callTool(t, session, "set_parameter", map[string]any{"nodeId": arr.NodeId, "value": "[1,2,3]"}, &set)
			require.True(t, set.Updated)

			number := map[string]any{"value": "2"}
			array := map[string]any{"nodeId": arr.NodeId, "port": "Value"}
			elems := []any{number, array}
			if arrayFirst {
				elems = []any{array, number}
			}

			var made polyformmcp.CreateNodesOutput
			callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
				map[string]any{"alias": "add", "type": multiplyNodeType, "inputs": map[string]any{
					"Values": map[string]any{"elements": elems},
				}},
			}}, &made)

			assert.Empty(t, made.Errors, "both elements should wire whichever order they come in")
			assert.NotEmpty(t, made.Nodes["add"])
		})
	}
}
