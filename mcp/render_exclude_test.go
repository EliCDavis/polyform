package mcp_test

import (
	"path/filepath"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSphereModel(t *testing.T, session *mcpsdk.ClientSession, radius, x string) string {
	t.Helper()

	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": radius}},
	}, &sphere)

	var model polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfModelType,
		"inputs": map[string]any{
			"Mesh":        map[string]any{"nodeId": sphere.NodeId, "port": "Out"},
			"Translation": map[string]any{"value": `{"x":` + x + `,"y":0,"z":0}`},
		},
	}, &model)
	return model.NodeId
}

// exclude used to match only top-level Models entries, so excluding a part
// nested under another ModelNode's Children did nothing and the two
// renders came back identical - which reads as "this part isn't the
// problem", the opposite of the truth, in the one tool meant to isolate a
// misplaced part without touching the graph.
func TestRenderExcludeAppliesToNestedChildren(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	child := makeSphereModel(t, session, "0.6", "3")

	var parent polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfModelType,
		"inputs": map[string]any{
			"Children": map[string]any{
				"elements": []any{map[string]any{"nodeId": child, "port": "Out"}},
			},
		},
	}, &parent)

	body := makeSphereModel(t, session, "1", "0")

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{
			"Models": map[string]any{
				"elements": []any{
					map[string]any{"nodeId": body, "port": "Out"},
					map[string]any{"nodeId": parent.NodeId, "port": "Out"},
				},
			},
		},
	}, &manifest)

	dir := t.TempDir()

	var full polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest.NodeId, "outputPath": filepath.Join(dir, "full.png"),
	}, &full)

	var without polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": filepath.Join(dir, "without.png"),
		"exclude":    []any{child},
	}, &without)

	require.Greater(t, full.TriangleCount, 0)
	assert.Less(t, without.TriangleCount, full.TriangleCount,
		"excluding a nested child must actually remove its triangles")
}

// A typo'd or wrong-kind id used to be accepted silently, producing an
// identical render that reads as a real result.
func TestRenderExcludeRejectsAnIdThatIsNotAPart(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	body := makeSphereModel(t, session, "1", "0")

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{
			"Models": map[string]any{
				"elements": []any{map[string]any{"nodeId": body, "port": "Out"}},
			},
		},
	}, &manifest)

	msg := callToolExpectingError(t, session, "render_preview", map[string]any{
		"nodeId":     manifest.NodeId,
		"outputPath": filepath.Join(t.TempDir(), "x.png"),
		"exclude":    []any{"Node-does-not-exist"},
	})
	assert.Contains(t, msg, "no node exists")
}
