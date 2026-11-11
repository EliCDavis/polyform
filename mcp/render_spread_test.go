package mcp_test

import (
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPreviewDrawsEveryModelSpreadIntoTheManifest(t *testing.T) {
	session := testSession(t)
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
		map[string]any{"alias": "a", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"A"`}, "Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "b", "type": gltfModelType, "inputs": map[string]any{
			"Name": map[string]any{"value": `"B"`}, "Mesh": map[string]any{"nodeId": "cube", "port": "Out"},
		}},
		map[string]any{"alias": "models", "type": fromElementsNodeType, "inputs": map[string]any{
			"Elements": map[string]any{"elements": []any{
				map[string]any{"nodeId": "a", "port": "Out"},
				map[string]any{"nodeId": "b", "port": "Out"},
			}},
		}},
		spreadOf("spread", "models"),
		map[string]any{"alias": "manifest", "type": gltfManifestType, "inputs": map[string]any{
			"Models": map[string]any{"elements": []any{map[string]any{"nodeId": "spread", "port": "Out"}}},
		}},
	}}, &created)
	require.Empty(t, created.Errors)

	var rendered polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": filepath.Join(t.TempDir(), "spread.png"),
	}, &rendered)
	assert.Equal(t, 24, rendered.TriangleCount, "both spread cubes")

	var partial polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": filepath.Join(t.TempDir(), "spread-only-b.png"),
		"exclude": []string{created.Nodes["a"]},
	}, &partial)
	assert.Equal(t, 12, partial.TriangleCount, "excluding a model that arrived through the spread")

	var framed polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": created.Nodes["manifest"], "outputPath": filepath.Join(t.TempDir(), "spread-framed.png"),
		"frameOn": []string{created.Nodes["spread"]},
	}, &framed)
	assert.Equal(t, 24, framed.TriangleCount, "the spread node itself names every model it hands over")
}
