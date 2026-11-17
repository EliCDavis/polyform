package mcp_test

import (
	"context"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestCreateVertexColorGradientSubgraph(t *testing.T) {
	session := testSession(t)

	var created polyformmcp.CreateVertexColorGradientSubgraphOutput
	callTool(t, session, "create_vertex_color_gradient_subgraph", map[string]any{
		"id":   "coat",
		"axis": "y",
	}, &created)
	require.Equal(t, "coat", created.SubgraphId)
	require.Equal(t, []string{"Mesh", "Color A", "Color B"}, created.Inputs)
	require.Equal(t, "Mesh", created.Output)

	// A default 1x1x1 cube - its own Y extent should drive the gradient's
	// range, not a hand-picked one.
	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType}, &cube)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "coat",
		"inputs": map[string]any{
			"Mesh":    map[string]any{"nodeId": cube.NodeId, "port": "Out"},
			"Color A": map[string]any{"value": `"#000000"`},
			"Color B": map[string]any{"value": `"#ffffff"`},
		},
	}, &instantiated)
	require.NotEmpty(t, instantiated.NodeId)
}

func TestCreateVertexColorGradientSubgraphRejectsBadAxis(t *testing.T) {
	session := testSession(t)

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "create_vertex_color_gradient_subgraph",
		Arguments: map[string]any{
			"id":   "bad",
			"axis": "W",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError, "axis must be X/Y/Z and should be rejected otherwise")
}
