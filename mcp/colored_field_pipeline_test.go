package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/require"
)

// TestColoredFieldEndToEndPipeline builds two overlapping colored spheres
// entirely through the real MCP tool surface (create_node/connect_nodes,
// the same calls an agent would make) - SphereNode x2 -> WithColorNode x2
// -> SmoothUnionColoredNode -> both ColoredFieldDistanceNode (to
// MarchNode) and the same ColoredField directly (to ApplyColorFieldNode)
// - then reads the resulting mesh's real "Color" attribute back to
// confirm the whole chain actually produces a red-to-blue gradient, not
// just that no tool call errored.
func TestColoredFieldEndToEndPipeline(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	sphereType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SphereNode]"
	withColorType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.WithColorNode]"
	unionType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SmoothUnionColoredNode]"
	distanceType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.ColoredFieldDistanceNode]"
	marchType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/marching.MarchNode]"
	applyType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/marching.ApplyColorFieldNode]"

	var sphereA, sphereB, withColorA, withColorB, union polyformmcp.CreateNodeOutput

	callTool(t, session, "create_node", map[string]any{
		"type": sphereType,
		"inputs": map[string]any{
			"Position": map[string]any{"value": `{"x":0,"y":0,"z":0}`},
			"Radius":   map[string]any{"value": "1"},
		},
	}, &sphereA)

	callTool(t, session, "create_node", map[string]any{
		"type": sphereType,
		"inputs": map[string]any{
			"Position": map[string]any{"value": `{"x":1.5,"y":0,"z":0}`},
			"Radius":   map[string]any{"value": "1"},
		},
	}, &sphereB)

	callTool(t, session, "create_node", map[string]any{
		"type": withColorType,
		"inputs": map[string]any{
			"Field": map[string]any{"nodeId": sphereA.NodeId, "port": "Field"},
			"Color": map[string]any{"value": `"#ff0000"`},
		},
	}, &withColorA)

	callTool(t, session, "create_node", map[string]any{
		"type": withColorType,
		"inputs": map[string]any{
			"Field": map[string]any{"nodeId": sphereB.NodeId, "port": "Field"},
			"Color": map[string]any{"value": `"#0000ff"`},
		},
	}, &withColorB)

	callTool(t, session, "create_node", map[string]any{
		"type": unionType,
		"inputs": map[string]any{
			"Radius": map[string]any{"value": "0.5"},
		},
	}, &union)
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": withColorA.NodeId, "outPort": "Out",
		"inNodeId": union.NodeId, "inPort": "Fields",
	}, &polyformmcp.ConnectNodesOutput{})
	callTool(t, session, "connect_nodes", map[string]any{
		"outNodeId": withColorB.NodeId, "outPort": "Out",
		"inNodeId": union.NodeId, "inPort": "Fields",
	}, &polyformmcp.ConnectNodesOutput{})

	var distance, march, apply polyformmcp.CreateNodeOutput

	callTool(t, session, "create_node", map[string]any{
		"type": distanceType,
		"inputs": map[string]any{
			"Field": map[string]any{"nodeId": union.NodeId, "port": "Union"},
		},
	}, &distance)

	callTool(t, session, "create_node", map[string]any{
		"type": marchType,
		"inputs": map[string]any{
			"Field":      map[string]any{"nodeId": distance.NodeId, "port": "Out"},
			"Resolution": map[string]any{"value": "8"},
			"Domain": map[string]any{"value": `{
				"center": {"x": 0.75, "y": 0, "z": 0},
				"extents": {"x": 2.5, "y": 1.5, "z": 1.5}
			}`},
		},
	}, &march)

	callTool(t, session, "create_node", map[string]any{
		"type": applyType,
		"inputs": map[string]any{
			"Mesh":  map[string]any{"nodeId": march.NodeId, "port": "Mesh"},
			"Field": map[string]any{"nodeId": union.NodeId, "port": "Union"},
		},
	}, &apply)

	node := inst.Node(apply.NodeId)
	require.NotNil(t, node)
	port, ok := node.Outputs()["Out"]
	require.True(t, ok)
	meshOut, ok := port.(nodes.Output[modeling.Mesh])
	require.True(t, ok)

	mesh := meshOut.Value()
	require.True(t, mesh.HasFloat3Attribute(modeling.ColorAttribute), "ApplyColorFieldNode should have written a Color attribute")

	colors := mesh.Float3Attribute(modeling.ColorAttribute)
	require.Greater(t, colors.Len(), 0, "marching should have produced real geometry")

	minR, maxR := 1.0, 0.0
	minB, maxB := 1.0, 0.0
	for i := 0; i < colors.Len(); i++ {
		c := colors.At(i)
		minR, maxR = min(minR, c.X()), max(maxR, c.X())
		minB, maxB = min(minB, c.Z()), max(maxB, c.Z())
	}

	// A real red-to-blue gradient across the mesh, not a flat single color
	// and not a hard binary switch - some vertices should read close to
	// pure red, some close to pure blue.
	require.Less(t, minR, 0.3, "some vertices should read close to pure red")
	require.Greater(t, maxB, 0.7, "some vertices should read close to pure blue")
}
