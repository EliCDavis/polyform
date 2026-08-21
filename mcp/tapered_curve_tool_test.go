package mcp_test

import (
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/require"
)

func TestCreateTaperedCurveSubgraph(t *testing.T) {
	session := testSession(t)

	var created polyformmcp.CreateTaperedCurveSubgraphOutput
	callTool(t, session, "create_tapered_curve_subgraph", map[string]any{
		"id": "tail",
	}, &created)
	require.Equal(t, "tail", created.SubgraphId)
	require.Equal(t, []string{"Points", "Base Radius", "Tip Radius", "Samples"}, created.Inputs)
	require.Equal(t, "Field", created.Output)

	// A gentle L-shaped bend through 3 sparse control points - exactly the
	// kind of hand-placed set that used to get wired straight into
	// VaryingRadiusLinesNode and come out faceted.
	var instantiated polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "tail",
		"inputs": map[string]any{
			"Points": map[string]any{
				"value": `[{"x":0,"y":0,"z":0},{"x":0,"y":1,"z":0},{"x":1,"y":1.5,"z":0}]`,
			},
			"Base Radius": map[string]any{"value": "0.2"},
			"Tip Radius":  map[string]any{"value": "0.05"},
			"Samples":     map[string]any{"value": "16"},
		},
	}, &instantiated)
	require.NotEmpty(t, instantiated.NodeId)

	var sampled polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"nodeId": instantiated.NodeId,
		"points": []map[string]any{
			// Exactly on the curve's start (a real control point, since a
			// Catmull-Rom spline always passes through every given point) -
			// distance from the surface should be ~ -Base Radius.
			{"x": 0, "y": 0, "z": 0},
			// Exactly on the curve's end - distance should be ~ -Tip Radius.
			{"x": 1, "y": 1.5, "z": 0},
			// Far from every control point - clearly outside.
			{"x": 10, "y": 10, "z": 10},
		},
	}, &sampled)
	require.Len(t, sampled.Values, 3)

	require.InDelta(t, -0.2, sampled.Values[0], 0.01, "distance at the curve's start should match Base Radius")
	require.InDelta(t, -0.05, sampled.Values[1], 0.01, "distance at the curve's end should match Tip Radius")
	require.Greater(t, sampled.Values[2], 0.0, "a point far from the curve should be outside the field")
}
