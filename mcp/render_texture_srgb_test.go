package mcp_test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A hex baked into a ColorTexture and the same hex run through
// SRGBToLinearNode into the BaseColorFactor describe the same surface, so
// they have to render the same. Sampling the texture without decoding it
// left the textured half visibly lighter, which had a whole pass tuning
// wood against a washed out picture.
func TestRenderPreviewDecodesColorTextureFromSRGB(t *testing.T) {
	session := testSession(t)

	const hex = `"#8a5a30"`
	uniformType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/texturing.UniformNode[github.com/EliCDavis/polyform/drawing/coloring.Color]]"
	toImageType := "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/texturing.ColorToImageNode]"

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
		}},

		map[string]any{"alias": "uniform", "type": uniformType, "inputs": map[string]any{
			"Fill": map[string]any{"value": hex}, "Width": map[string]any{"value": "8"}, "Height": map[string]any{"value": "8"},
		}},
		map[string]any{"alias": "image", "type": toImageType, "inputs": map[string]any{
			"Texture": map[string]any{"nodeId": "uniform", "port": "Texture"},
		}},
		map[string]any{"alias": "texture", "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.TextureNode]",
			"inputs": map[string]any{"Image": map[string]any{"nodeId": "image", "port": "Image"}}},
		map[string]any{"alias": "texturedMat", "type": materialNodeType, "inputs": map[string]any{
			"Color Texture": map[string]any{"nodeId": "texture", "port": "Out"},
		}},
		map[string]any{"alias": "texturedModel", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"}, "Material": map[string]any{"nodeId": "texturedMat", "port": "Out"},
		}},

		map[string]any{"alias": "linear", "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/drawing/coloring.SRGBToLinearNode]",
			"inputs": map[string]any{"In": map[string]any{"value": hex}}},
		map[string]any{"alias": "flatMat", "type": materialNodeType, "inputs": map[string]any{
			"Color": map[string]any{"nodeId": "linear", "port": "Out"},
		}},
		map[string]any{"alias": "flatModel", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"}, "Material": map[string]any{"nodeId": "flatMat", "port": "Out"},
		}},

		map[string]any{"alias": "texturedManifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "texturedModel", "port": "Out"},
		}},
		map[string]any{"alias": "flatManifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "flatModel", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	dir := t.TempDir()
	centerPixel := func(manifest, name string) (uint32, uint32, uint32) {
		path := filepath.Join(dir, name)
		callTool(t, session, "render_preview", map[string]any{
			"nodeId": created.Nodes[manifest], "outputPath": path,
			"width": 64, "height": 64,
			"views": []map[string]any{{"azimuth": 0, "elevation": 0}},
		}, &polyformmcp.RenderPreviewOutput{})

		f, err := os.Open(path)
		require.NoError(t, err)
		defer f.Close()
		img, _, err := image.Decode(f)
		require.NoError(t, err)
		r, g, b, _ := img.At(32, 32).RGBA()
		return r >> 8, g >> 8, b >> 8
	}

	tr, tg, tb := centerPixel("texturedManifest", "textured.png")
	fr, fg, fb := centerPixel("flatManifest", "flat.png")

	// The texture is 8-bit sRGB and the factor is full-precision float, so
	// they agree only to within a quantization step - widest in the darkest
	// channel. Sampling without the decode put them ~60 apart.
	assert.InDelta(t, fr, tr, 4, "red: textured %d vs factor %d", tr, fr)
	assert.InDelta(t, fg, tg, 4, "green: textured %d vs factor %d", tg, fg)
	assert.InDelta(t, fb, tb, 4, "blue: textured %d vs factor %d", tb, fb)
}
