package mcp_test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

const (
	materialNodeType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.MaterialNode]"
	gltfModelType    = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ModelNode]"
	gltfManifestType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.ManifestNode]"
	setAttribute3D   = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.SetAttribute3DNode]"
)

// renderWhiteVertexColoredCube builds a cube whose every vertex Color is
// pure white, optionally under a material of the given base color, and
// returns the rendered center pixel. White vertex colors are the identity
// for a multiply, so whatever comes back is exactly what the preview did
// with the material's factor.
func renderWhiteVertexColoredCube(t *testing.T, session *mcpsdk.ClientSession, baseColor string) (uint32, uint32, uint32) {
	t.Helper()

	nodes := []any{
		map[string]any{
			"alias": "cube",
			"type":  cubeNodeType,
			"inputs": map[string]any{
				"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
			},
		},
		map[string]any{
			"alias": "select",
			"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.SelectFromMeshNode]",
			"inputs": map[string]any{
				"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
			},
		},
		map[string]any{
			"alias": "axis",
			"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/vector3.SelectArray[float64]]",
			"inputs": map[string]any{
				"In": map[string]any{"nodeId": "select", "port": "Position"},
			},
		},
		map[string]any{
			"alias": "time",
			"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math.RemapToArrayNode[float64]]",
			"inputs": map[string]any{
				"Value":  map[string]any{"nodeId": "axis", "port": "Y"},
				"InMin":  map[string]any{"value": "-1"},
				"InMax":  map[string]any{"value": "1"},
				"OutMin": map[string]any{"value": "0"},
				"OutMax": map[string]any{"value": "1"},
			},
		},
		// Both endpoints white, so every vertex Color is exactly (1,1,1).
		map[string]any{
			"alias": "interpolate",
			"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.InterpolateToArrayNode]",
			"inputs": map[string]any{
				"A":    map[string]any{"value": `"#ffffff"`},
				"B":    map[string]any{"value": `"#ffffff"`},
				"Time": map[string]any{"nodeId": "time", "port": "Out"},
			},
		},
		map[string]any{
			"alias": "white",
			"type":  "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.ToVectorArrayNode]",
			"inputs": map[string]any{
				"In": map[string]any{"nodeId": "interpolate", "port": "Out"},
			},
		},
		map[string]any{
			"alias": "colored",
			"type":  setAttribute3D,
			"inputs": map[string]any{
				"Mesh":      map[string]any{"nodeId": "cube", "port": "Out"},
				"Attribute": map[string]any{"value": `"Color"`},
				"Data":      map[string]any{"nodeId": "white", "port": "Vector 3"},
			},
		},
		map[string]any{
			"alias": "model",
			"type":  gltfModelType,
			"inputs": map[string]any{
				"Mesh": map[string]any{"nodeId": "colored", "port": "Out"},
			},
		},
		map[string]any{
			"alias": "manifest",
			"type":  gltfManifestType,
			"inputs": map[string]any{
				"Models": map[string]any{"nodeId": "model", "port": "Out"},
			},
		},
	}

	if baseColor != "" {
		nodes = append(nodes, map[string]any{
			"alias":  "material",
			"type":   materialNodeType,
			"inputs": map[string]any{"Color": map[string]any{"value": baseColor}},
		})
		for _, entry := range nodes {
			if entry.(map[string]any)["alias"] == "model" {
				entry.(map[string]any)["inputs"].(map[string]any)["Material"] = map[string]any{"nodeId": "material", "port": "Out"}
			}
		}
	}

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": nodes}, &created)
	require.Empty(t, created.Errors)

	outPath := filepath.Join(t.TempDir(), "tint.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     created.Nodes["manifest"],
		"outputPath": outPath,
		"views":      []map[string]any{{"azimuth": 0, "elevation": 0}},
	}, &out)

	f, err := os.Open(outPath)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	bounds := img.Bounds()
	r, g, b, _ := img.At(bounds.Dx()/2, bounds.Dy()/2).RGBA()
	return r, g, b
}

// glTF multiplies COLOR_0 by baseColorFactor. The preview used to let
// vertex color win outright, so a mesh carrying subtle near-white
// variation over a colored material rendered white here and correctly
// tinted in every real viewer - and the fix for that disagreement was to
// bake the material color into the vertex data, which is not something a
// caller should have to do.
func TestRenderPreviewMultipliesVertexColorByMaterialColor(t *testing.T) {
	session := testSession(t)

	r, g, b := renderWhiteVertexColoredCube(t, session, `"#ff0000"`)

	require.Greater(t, r, g, "a red material must tint white vertex colors red")
	require.Greater(t, r, b, "a red material must tint white vertex colors red")
}

// A material that sets no base color factor must not darken a mesh that
// brought its own colors: the neutral gray stand-in used for flat shading
// is a preview convenience, not a real factor to multiply by.
func TestRenderPreviewDoesNotTintVertexColorsWithoutAMaterial(t *testing.T) {
	session := testSession(t)

	r, g, b := renderWhiteVertexColoredCube(t, session, "")

	// Pure white vertex colors, lit: the face should read as near-neutral
	// and bright, not knocked down toward the 0.7 gray stand-in.
	require.InDelta(t, r, g, 3000, "no material should leave vertex colors neutral")
	require.InDelta(t, r, b, 3000, "no material should leave vertex colors neutral")
	require.Greater(t, r, uint32(30000), "white vertex colors should stay bright")
}
