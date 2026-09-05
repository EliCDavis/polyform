package mcp_test

import (
	"math"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sdfSphereNodeType      = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SphereNode]"
	sdfSmoothUnionNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SmoothUnionNode]"
	sdfSubtractNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.SubtractionNode]"
)

func vec(x, y, z float64) map[string]any {
	return map[string]any{"x": x, "y": y, "z": z}
}

// A lone sphere is the case with an exact answer to check against.
func TestRaycastFieldFindsAKnownSurface(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sdfSphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "1.5"}},
	}, &sphere)

	var out polyformmcp.RaycastFieldOutput
	callTool(t, session, "raycast_field", map[string]any{
		"nodeId": sphere.NodeId,
		"rays": []any{
			// From outside, inward.
			map[string]any{"origin": vec(5, 0, 0), "direction": vec(-1, 0, 0)},
			// From the centre, outward - attaching to a cavity wall works
			// this way round.
			map[string]any{"origin": vec(0, 0, 0), "direction": vec(0, 1, 0)},
		},
	}, &out)

	require.Len(t, out.Hits, 2)

	outside := out.Hits[0]
	require.True(t, outside.Hit)
	assert.InDelta(t, 1.5, outside.Point.X, 0.001)
	assert.InDelta(t, 3.5, outside.Distance, 0.001)
	assert.InDelta(t, 1, outside.Normal.X, 0.01, "normal points out of the sphere")

	inside := out.Hits[1]
	require.True(t, inside.Hit, "marching outward from inside must still find the surface")
	assert.InDelta(t, 1.5, inside.Point.Y, 0.001)
	assert.InDelta(t, 1, inside.Normal.Y, 0.01)
}

// The case that actually matters: a mouth cavity subtracted out of a
// blended two-lobe body. There is no formula for where that surface sits,
// which is why builds hand-tuned tooth coordinates by render feedback and
// shipped teeth floating off the snout.
func TestRaycastFieldFindsACavityWallInABlendedBody(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "head", "type": sdfSphereNodeType,
				"inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
			map[string]any{"alias": "body", "type": sdfSphereNodeType,
				"inputs": map[string]any{
					"Radius":   map[string]any{"value": "0.9"},
					"Position": map[string]any{"value": `{"x":0,"y":0,"z":-1.2}`},
				}},
			map[string]any{"alias": "blend", "type": sdfSmoothUnionNodeType,
				"inputs": map[string]any{
					"Radius": map[string]any{"value": "0.3"},
					"Fields": map[string]any{"elements": []any{
						map[string]any{"nodeId": "head", "port": "Field"},
						map[string]any{"nodeId": "body", "port": "Field"},
					}},
				}},
			map[string]any{"alias": "cavity", "type": sdfSphereNodeType,
				"inputs": map[string]any{
					"Radius":   map[string]any{"value": "0.7"},
					"Position": map[string]any{"value": `{"x":0,"y":0,"z":0.55}`},
				}},
			map[string]any{"alias": "mouth", "type": sdfSubtractNodeType,
				"inputs": map[string]any{
					"A": map[string]any{"nodeId": "blend", "port": "Union"},
					"B": map[string]any{"nodeId": "cavity", "port": "Field"},
				}},
		},
	}, &out)
	require.Empty(t, out.Errors)

	// March up from inside the cavity to find its roof - where an upper
	// tooth row has to sit.
	var hit polyformmcp.RaycastFieldOutput
	callTool(t, session, "raycast_field", map[string]any{
		"nodeId": out.Nodes["mouth"],
		"rays": []any{
			map[string]any{"origin": vec(0, 0, 0.55), "direction": vec(0, 1, 0)},
		},
	}, &hit)

	require.Len(t, hit.Hits, 1)
	roof := hit.Hits[0]
	require.True(t, roof.Hit, "should have found the cavity roof")

	// It really is on the surface: the field is ~0 there.
	var samples polyformmcp.SampleFieldOutput
	callTool(t, session, "sample_field", map[string]any{
		"nodeId": out.Nodes["mouth"],
		"points": []any{vec(roof.Point.X, roof.Point.Y, roof.Point.Z)},
	}, &samples)
	require.Len(t, samples.Samples, 1)
	assert.InDelta(t, 0, samples.Samples[0].Value, 0.01,
		"the reported point must lie on the surface, not near it")

	// And the normal is a unit vector, so it can be used to orient and
	// offset a part directly.
	n := math.Sqrt(roof.Normal.X*roof.Normal.X + roof.Normal.Y*roof.Normal.Y + roof.Normal.Z*roof.Normal.Z)
	assert.InDelta(t, 1, n, 0.01)
}

// A ray pointing into empty space has to say so rather than reporting a
// point that isn't on anything.
func TestRaycastFieldReportsAMiss(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sdfSphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "1"}},
	}, &sphere)

	var out polyformmcp.RaycastFieldOutput
	callTool(t, session, "raycast_field", map[string]any{
		"nodeId": sphere.NodeId,
		"rays": []any{
			map[string]any{"origin": vec(5, 0, 0), "direction": vec(1, 0, 0)},
		},
	}, &out)

	require.Len(t, out.Hits, 1)
	assert.False(t, out.Hits[0].Hit)
	assert.Contains(t, out.Hits[0].Note, "no surface")
}
