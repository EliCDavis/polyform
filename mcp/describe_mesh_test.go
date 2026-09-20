package mcp_test

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sdfUnionNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/sdf.UnionNode]"
	marchNodeType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/marching.MarchNode]"
)

// marchTwoSpheres builds two spheres unioned and marched, `gap` apart on
// X. At a small gap they fuse into one body; at a large one they don't.
func marchTwoSpheres(t *testing.T, session *mcpsdk.ClientSession, gap string) string {
	t.Helper()

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "a", "type": sdfSphereNodeType,
				"inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
			map[string]any{"alias": "b", "type": sdfSphereNodeType,
				"inputs": map[string]any{
					"Radius":   map[string]any{"value": "0.5"},
					"Position": map[string]any{"value": `{"x":` + gap + `,"y":0,"z":0}`},
				}},
			map[string]any{"alias": "union", "type": sdfUnionNodeType,
				"inputs": map[string]any{
					"Fields": map[string]any{"elements": []any{
						map[string]any{"nodeId": "a", "port": "Field"},
						map[string]any{"nodeId": "b", "port": "Field"},
					}},
				}},
			map[string]any{"alias": "march", "type": marchNodeType,
				"inputs": map[string]any{
					"Field":      map[string]any{"nodeId": "union", "port": "Union"},
					"Resolution": map[string]any{"value": "40"},
					"Domain":     map[string]any{"value": `{"center":{"x":0,"y":0,"z":0},"extents":{"x":8,"y":4,"z":4}}`},
				}},
		},
	}, &out)
	require.Empty(t, out.Errors)
	return out.Nodes["march"]
}

// The case a render is worst at. Two shapes that were meant to fuse but
// didn't look completely fine from most angles - a tooth floating just off
// a jaw is only visible edge-on - while the piece count says so outright.
func TestDescribeMeshCountsDetachedPieces(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	fused := marchTwoSpheres(t, session, "0.9")
	var joined polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": fused}, &joined)

	require.Greater(t, joined.Triangles, 100)
	assert.Equal(t, 1, joined.Pieces, "overlapping spheres should march as one body")
}

func TestDescribeMeshSeesAPartThatCameAdrift(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	adrift := marchTwoSpheres(t, session, "2.5")
	var split polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": adrift}, &split)

	assert.Equal(t, 2, split.Pieces, "a part that didn't reach the body is its own piece")
	require.Len(t, split.LargestFirst, 2)

	// Biggest first, so the stray is last and its position is reported -
	// which is what tells you which part came adrift.
	assert.Greater(t, split.LargestFirst[0].Triangles, split.LargestFirst[1].Triangles)
	assert.InDelta(t, 2.5, split.LargestFirst[1].Center.X, 0.2)
}

// Marching emits a separate vertex per triangle corner, so grouping by
// vertex index would report one piece per triangle. Positions get welded
// instead; this pins that the welding actually happens.
func TestDescribeMeshWeldsSeamVerticesBeforeCounting(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	one := marchTwoSpheres(t, session, "0.9")
	var d polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": one}, &d)

	require.Greater(t, d.Triangles, 100)
	assert.Equal(t, 1, d.Pieces,
		"got %d pieces from %d triangles - seam vertices were not welded", d.Pieces, d.Triangles)
}

// The proportion check, without a render.
func TestDescribeMeshReportsExtentForProportionChecks(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	m := marchTwoSpheres(t, session, "0.9")
	var d polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": m}, &d)

	// A radius-1 sphere plus a radius-0.5 one centred 0.9 away spans about
	// -1 to 1.4 on X, and -1 to 1 on the other two.
	assert.InDelta(t, 2.4, d.Extent.Size.X, 0.25)
	assert.InDelta(t, 2.0, d.Extent.Size.Y, 0.25)
	assert.Greater(t, d.Extent.Size.X, d.Extent.Size.Y, "it is longer than it is tall")
}

// An empty mesh is the "did my part disappear" case, and it should say the
// usual causes rather than just reporting zero.
func TestDescribeMeshExplainsAnEmptyMesh(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"nodes": []any{
			map[string]any{"alias": "s", "type": sdfSphereNodeType,
				"inputs": map[string]any{"Radius": map[string]any{"value": "1"}}},
			map[string]any{"alias": "march", "type": marchNodeType,
				"inputs": map[string]any{
					"Field":      map[string]any{"nodeId": "s", "port": "Field"},
					"Resolution": map[string]any{"value": "40"},
					// Nowhere near the sphere.
					"Domain": map[string]any{"value": `{"center":{"x":50,"y":0,"z":0},"extents":{"x":2,"y":2,"z":2}}`},
				}},
		},
	}, &out)
	require.Empty(t, out.Errors)

	var d polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{"nodeId": out.Nodes["march"]}, &d)

	assert.Equal(t, 0, d.Triangles)
	assert.Contains(t, d.Note, "empty")
	assert.Contains(t, d.Note, "Domain")
}

// It shares the subgraph-scope resolution, so it must carry the same
// caveat rather than silently reporting definition defaults.
func TestDescribeMeshWarnsWhenProbingADefinition(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	_, instanceID := buildScopedSphere(t, session, "3")

	var out polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{
		"scope": "part",
		"nodes": []any{
			map[string]any{"alias": "march", "type": marchNodeType,
				"inputs": map[string]any{"Resolution": map[string]any{"value": "20"}}},
		},
	}, &out)
	require.Empty(t, out.Errors)

	var d polyformmcp.DescribeMeshOutput
	callTool(t, session, "describe_mesh", map[string]any{
		"scope": "part", "nodeId": out.Nodes["march"],
	}, &d)

	require.NotEmpty(t, d.EvaluationContext)
	assert.Contains(t, d.EvaluationContext, instanceID)
}
