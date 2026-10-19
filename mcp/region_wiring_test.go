package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	areaNodeType             = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/extrude.AreaNode]"
	fillNodeType             = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.FillNode]"
	unionAreaNodeType        = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.UnionNode]"
	roundedRectangleNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.RoundedRectangleNode]"
	transformAreaNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/sketch.TransformAreaNode]"
)

func TestAnArrayLiteralFeedsALiftedPortAsAnArray(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trs.NewNode]",
		"inputs": map[string]any{
			"Position": map[string]any{"value": `[{"x":1,"y":0,"z":0},{"x":2,"y":0,"z":0}]`},
		},
	}, &created)

	var out polyformmcp.EvaluateNodeOutput
	callTool(t, session, "evaluate_node", map[string]any{"nodeId": created.NodeId}, &out)
	assert.Equal(t, 2, out.Length, "one transform per position")
}

func TestRegionTakesOneOutlineAndASpreadListOfHoles(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "board", "type": roundedRectangleNodeType, "inputs": map[string]any{
			"Size":   map[string]any{"value": `{"x":10,"y":6}`},
			"Radius": map[string]any{"value": "1"},
		}},
		map[string]any{"alias": "pads", "type": roundedRectangleNodeType, "inputs": map[string]any{
			"Center": map[string]any{"value": `[{"x":-2,"y":0},{"x":2,"y":0}]`},
		}},
		spreadOf("spread", "pads"),
		map[string]any{"alias": "fill", "type": fillNodeType, "inputs": map[string]any{
			"Outlines": map[string]any{"elements": []any{map[string]any{"nodeId": "board", "port": "Out"}}},
			"Holes":    map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
		map[string]any{"alias": "region", "type": areaNodeType, "inputs": map[string]any{
			"Area":  map[string]any{"nodeId": "fill", "port": "Out"},
			"Depth": map[string]any{"value": "0.5"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	msg := callToolExpectingError(t, session, "connect_nodes", map[string]any{
		"outNodeId": created.Nodes["pads"], "outPort": "Out", "inNodeId": created.Nodes["fill"], "inPort": "Holes",
	})
	assert.Contains(t, msg, `"Holes" input takes`, "a list wired straight into one element is refused")

	var mesh polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": created.Nodes["region"]}, &mesh)
	assert.Equal(t, 1, mesh.Pieces)
	assert.InDelta(t, 10, mesh.Extent.Size.X, 1e-9)
	assert.InDelta(t, 0.5, mesh.Extent.Size.Z, 1e-9)
	assert.Empty(t, mesh.NodeErrors)
}

func TestRegionTakesASpreadListOfStrokedRings(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "seam", "type": strokeNodeType, "inputs": map[string]any{
			"Path":   map[string]any{"value": `[{"x":0,"y":0},{"x":4,"y":0},{"x":4,"y":4},{"x":0,"y":4}]`},
			"Width":  map[string]any{"value": "0.5"},
			"Closed": map[string]any{"value": "true"},
		}},
		map[string]any{"alias": "seams", "type": transformAreaNodeType, "inputs": map[string]any{
			"Area":        map[string]any{"nodeId": "seam", "port": "Out"},
			"Translation": map[string]any{"value": `[{"x":0,"y":0},{"x":10,"y":0}]`},
		}},
		spreadOf("spread", "seams"),
		map[string]any{"alias": "union", "type": unionAreaNodeType, "inputs": map[string]any{
			"Areas": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
		map[string]any{"alias": "region", "type": areaNodeType, "inputs": map[string]any{
			"Area":  map[string]any{"nodeId": "union", "port": "Out"},
			"Depth": map[string]any{"value": "0.1"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var mesh polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": created.Nodes["region"]}, &mesh)
	assert.Equal(t, 2, mesh.Pieces, "two rings, each a single piece")
	assert.Empty(t, mesh.NodeErrors)
}
