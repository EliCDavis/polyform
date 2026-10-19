package mcp_test

import (
	"strings"
	"testing"

	_ "github.com/EliCDavis/polyform/math/vector2"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	vector2NewNodeType  = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector2.NewNode[float64]]"
	gltfTextureNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.TextureNode]"
)

func TestConnectNodesBatchKeepsItsEdgesWhenAnotherNoLongerFits(t *testing.T) {
	session, inst := testSessionWithInstance(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "size", "type": vector2NewNodeType, "inputs": map[string]any{"X": map[string]any{"value": "1"}, "Y": map[string]any{"value": "1"}}},
		map[string]any{"alias": "texture", "type": gltfTextureNodeType, "inputs": map[string]any{"Offset": map[string]any{"nodeId": "size", "port": "Out"}}},
		map[string]any{"alias": "widths", "type": floatArrayParamType, "value": "[1,2]"},
		map[string]any{"alias": "three", "type": floatParamType, "value": "3"},
		map[string]any{"alias": "product", "type": multiplyNodeType},
	}}, &made)
	require.Empty(t, made.Errors)

	var out polyformmcp.ConnectNodesOutput
	callTool(t, session, "connect_nodes", map[string]any{"connections": []any{
		map[string]any{"outNodeId": made.Nodes["widths"], "outPort": "Value", "inNodeId": made.Nodes["size"], "inPort": "X"},
		map[string]any{"outNodeId": made.Nodes["three"], "outPort": "Value", "inNodeId": made.Nodes["product"], "inPort": "Values"},
	}}, &out)

	assert.False(t, out.Connected)
	assert.Contains(t, strings.Join(out.Errors, "\n"), "removed")
	assert.InDelta(t, 3, evalFloat64Output(t, inst, made.Nodes["product"], "Float"), 1e-9, "the edge that fits is kept")
}
