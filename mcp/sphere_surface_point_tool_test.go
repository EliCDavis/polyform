package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/require"
)

func TestCreateSphereSurfacePointSubgraph(t *testing.T) {
	session, inst := testSessionWithInstance(t)

	var created polyformmcp.CreateSphereSurfacePointSubgraphOutput
	callTool(t, session, "create_sphere_surface_point_subgraph", map[string]any{
		"id": "eye_pos",
	}, &created)
	require.Equal(t, "eye_pos", created.SubgraphId)
	require.Equal(t, []string{"Center", "Radius", "Direction", "Embed Fraction"}, created.Inputs)
	require.Equal(t, "Position", created.Output)

	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "eye_pos",
		"inputs": map[string]any{
			"Center": map[string]any{"value": `{"x":1,"y":2,"z":3}`},
			"Radius": map[string]any{"value": "1"},
			// Deliberately non-unit-length, to prove normalization happens.
			"Direction":      map[string]any{"value": `{"x":0,"y":4,"z":0}`},
			"Embed Fraction": map[string]any{"value": "0.5"},
		},
	}, &instantiated)
	require.NotEmpty(t, instantiated.NodeId)

	node := inst.Node(instantiated.NodeId)
	require.NotNil(t, node)
	port, ok := node.Outputs()["Position"]
	require.True(t, ok)
	valued, ok := port.(nodes.Output[vector3.Float64])
	require.True(t, ok)

	got := valued.Value()
	want := vector3.New(1.0, 2.5, 3.0) // Center + (0,1,0)*Radius*0.5
	require.InDelta(t, want.X(), got.X(), 1e-9)
	require.InDelta(t, want.Y(), got.Y(), 1e-9)
	require.InDelta(t, want.Z(), got.Z(), 1e-9)
}
