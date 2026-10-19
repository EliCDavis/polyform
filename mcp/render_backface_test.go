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

const flipWindingType = "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling/meshops.FlipTriangleWindingNode]"

func TestRenderPreviewShowsAnInsideOutSurfaceInMagenta(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
		}},
		map[string]any{"alias": "flip", "type": flipWindingType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "flip", "port": "Flipped"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "model", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	center := func(showBackFaces bool) (uint32, uint32, uint32) {
		path := filepath.Join(t.TempDir(), "inside-out.png")
		callTool(t, session, "render_preview", map[string]any{
			"nodeId": created.Nodes["manifest"], "outputPath": path, "showBackFaces": showBackFaces,
			"views": []map[string]any{{"azimuth": 0, "elevation": 0}},
		}, &polyformmcp.RenderPreviewOutput{})
		f, err := os.Open(path)
		require.NoError(t, err)
		defer f.Close()
		img, _, err := image.Decode(f)
		require.NoError(t, err)
		r, g, b, _ := img.At(img.Bounds().Dx()/2, img.Bounds().Dy()/2).RGBA()
		return r, g, b
	}

	r, g, b := center(true)
	assert.Greater(t, r, uint32(60000))
	assert.Less(t, g, uint32(5000))
	assert.Greater(t, b, uint32(60000))

	_, g, _ = center(false)
	assert.Greater(t, g, uint32(5000), "culled as a viewer would, the far walls show through instead")
}
