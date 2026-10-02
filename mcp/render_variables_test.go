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

func TestRenderPreviewVariablesOverrideForOneRenderOnly(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Paint", "type": "coloring.color", "value": `"#0000ff"`}},
	}, &polyformmcp.CreateVariablesOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
		}},
		map[string]any{"alias": "material", "type": materialNodeType, "inputs": map[string]any{
			"Color": map[string]any{"variable": "Paint"},
		}},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Mesh":     map[string]any{"nodeId": "cube", "port": "Out"},
			"Material": map[string]any{"nodeId": "material", "port": "Out"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "model", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	centerPixel := func(path string) (uint32, uint32, uint32) {
		f, err := os.Open(path)
		require.NoError(t, err)
		defer f.Close()
		img, _, err := image.Decode(f)
		require.NoError(t, err)
		b := img.Bounds()
		r, g, bl, _ := img.At(b.Dx()/2, b.Dy()/2).RGBA()
		return r, g, bl
	}

	red := filepath.Join(t.TempDir(), "red.png")
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": red,
		"views":     []map[string]any{{"azimuth": 0, "elevation": 0}},
		"variables": map[string]any{"Paint": `"#ff0000"`},
	}, &polyformmcp.RenderPreviewOutput{})
	r, _, b := centerPixel(red)
	assert.Greater(t, r, b, "the override paints the cube red for this render")

	var list polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &list)
	require.Len(t, list.Variables, 1)
	assert.Equal(t, "#0000ff", list.Variables[0].Value, "the variable is restored afterward")

	blue := filepath.Join(t.TempDir(), "blue.png")
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": blue,
		"views": []map[string]any{{"azimuth": 0, "elevation": 0}},
	}, &polyformmcp.RenderPreviewOutput{})
	r, _, b = centerPixel(blue)
	assert.Greater(t, b, r, "a plain render sees the restored value")

	msg := callToolExpectingError(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": blue,
		"variables": map[string]any{"Nope": `1`},
	})
	assert.Contains(t, msg, "Nope")
}
