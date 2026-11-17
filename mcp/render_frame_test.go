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

func TestRenderViewFrameOwnFitsEachLookToItself(t *testing.T) {
	session := testSession(t)

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Size", "type": "float64", "value": "4"}},
	}, &polyformmcp.CreateVariablesOutput{})

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"variable": "Size"}, "Height": map[string]any{"variable": "Size"}, "Depth": map[string]any{"variable": "Size"},
		}},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "model", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	path := filepath.Join(t.TempDir(), "frames.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": path,
		"width": 64, "height": 64,
		"views": []map[string]any{
			{"variables": map[string]any{"Size": "0.5"}},
			{"variables": map[string]any{"Size": "0.5"}, "frame": "own"},
		},
	}, &out)
	require.Equal(t, 2, out.Columns)

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	background := img.At(0, 0)
	covered := func(x0, x1 int) int {
		n := 0
		for y := 0; y < 64; y++ {
			for x := x0; x < x1; x++ {
				if img.At(x, y) != background {
					n++
				}
			}
		}
		return n
	}
	shared, own := covered(0, 64), covered(64, 128)
	assert.Greater(t, own, shared*3, "own framing fills the view with the small cube; shared framing leaves it a speck (shared=%d own=%d)", shared, own)
}

func TestRenderViewFitRadiusFramesATargetRegardlessOfSceneSize(t *testing.T) {
	session := testSession(t)

	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"value": "10"}, "Height": map[string]any{"value": "10"}, "Depth": map[string]any{"value": "10"},
		}},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"nodeId": "model", "port": "Out"},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	path := filepath.Join(t.TempDir(), "fit.png")
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": path,
		"width": 64, "height": 64,
		"views": []map[string]any{
			{"target": map[string]any{"x": 0, "y": 0, "z": 5}},
			{"target": map[string]any{"x": 0, "y": 0, "z": 5}, "fitRadius": 0.5},
		},
	}, &polyformmcp.RenderPreviewOutput{})

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	background := img.At(0, 0)
	covered := func(x0 int) int {
		n := 0
		for y := 0; y < 64; y++ {
			for x := x0; x < x0+64; x++ {
				if img.At(x, y) != background {
					n++
				}
			}
		}
		return n
	}
	assert.Equal(t, 64*64, covered(64), "a 0.5 m sphere on the face of a 10 m cube is all cube")
	assert.Less(t, covered(0), 64*64, "scene-relative framing shows the cube's edges")
}

func TestRenderViewFrameRejectsUnknownValues(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{"Mesh": map[string]any{"nodeId": "cube", "port": "Out"}}},
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{"Models": map[string]any{"nodeId": "model", "port": "Out"}}},
	}}, &created)

	msg := callToolExpectingError(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": filepath.Join(t.TempDir(), "x.png"),
		"views": []map[string]any{{"frame": "tight"}},
	})
	assert.Contains(t, msg, "frame must be 'own'")
}
