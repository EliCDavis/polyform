package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/modeling/sketch"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	floatArrayParamType = "github.com/EliCDavis/polyform/generator/parameter.Value[[]float64]"
	strokeNodeType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.StrokeNode]"
	outlineNodeType     = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/extrude.OutlineNode]"
)

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

func TestALiftedPortOfPathsTakesAPathLiteral(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "one", "type": strokeNodeType, "inputs": map[string]any{
			"Path": map[string]any{"value": `[{"x":0,"y":0},{"x":1,"y":0}]`},
		}},
		map[string]any{"alias": "sweep", "type": outlineNodeType, "inputs": map[string]any{
			"Outline": map[string]any{"value": `[{"x":-1,"y":-1},{"x":1,"y":-1},{"x":1,"y":1},{"x":-1,"y":1}]`},
			"Path":    map[string]any{"value": `[{"x":0,"y":0,"z":0},{"x":0,"y":0,"z":5}]`},
		}},
	}}, &made)
	require.Empty(t, made.Errors)

	one := nodes.GetNodeOutputPort[sketch.Area](inst.Node(made.Nodes["one"]), "Out").Value()
	assert.Len(t, one.Polygons, 1)
}
