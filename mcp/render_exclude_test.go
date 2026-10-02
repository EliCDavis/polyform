package mcp_test

import (
	"image"
	_ "image/png"
	"os"
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

func TestRenderViewsTakeTheirOwnExcludeAndVariables(t *testing.T) {
	session, _ := testSessionWithInstance(t)

	callTool(t, session, "create_variables", map[string]any{
		"variables": []map[string]any{{"path": "Offset", "type": "vector3.Vector[float64]", "value": `{"x":3,"y":0,"z":0}`}},
	}, &polyformmcp.CreateVariablesOutput{})

	body := makeSphereModel(t, session, "1", "0")
	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   sphereNodeType,
		"inputs": map[string]any{"Radius": map[string]any{"value": "0.6"}},
	}, &sphere)
	var extra polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfModelType,
		"inputs": map[string]any{
			"Mesh":        map[string]any{"nodeId": sphere.NodeId, "port": "Out"},
			"Translation": map[string]any{"variable": "Offset"},
		},
	}, &extra)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{"Models": map[string]any{"elements": []any{
			map[string]any{"nodeId": body, "port": "Out"},
			map[string]any{"nodeId": extra.NodeId, "port": "Out"},
		}}},
	}, &manifest)

	outPath := filepath.Join(t.TempDir(), "views.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest.NodeId, "outputPath": outPath, "width": 64, "height": 64,
		"views": []map[string]any{
			{"name": "full", "azimuth": 0},
			{"name": "without", "azimuth": 0, "exclude": []any{extra.NodeId}},
			{"name": "moved", "azimuth": 0, "variables": map[string]any{"Offset": `{"x":-3,"y":0,"z":0}`}},
			{"name": "same", "azimuth": 0},
		},
	}, &out)
	require.Equal(t, 4, out.Views)

	f, err := os.Open(outPath)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)

	// Same framing in every cell, so cells differ only where the second
	// sphere was excluded or moved.
	cellW := img.Bounds().Dx() / out.Columns
	cellH := img.Bounds().Dy() / out.Rows
	differing := func(a, b int) int {
		ax, ay := (a%out.Columns)*cellW, (a/out.Columns)*cellH
		bx, by := (b%out.Columns)*cellW, (b/out.Columns)*cellH
		n := 0
		for y := 0; y < cellH-12; y++ {
			for x := 0; x < cellW; x++ {
				if img.At(ax+x, ay+y) != img.At(bx+x, by+y) {
					n++
				}
			}
		}
		return n
	}
	// The caption text differs between cells; the geometry must not.
	baseline := differing(0, 3)
	assert.Less(t, baseline, 40, "two views with no overrides render identically apart from their captions")
	assert.Greater(t, differing(0, 1), baseline+100, "the per-view exclude removed the second sphere")
	assert.Greater(t, differing(0, 2), baseline+100, "the per-view variable moved the second sphere")

	var offset polyformmcp.ListVariablesOutput
	callTool(t, session, "list_variables", map[string]any{}, &offset)
	for _, v := range offset.Variables {
		if v.Path == "Offset" {
			assert.Equal(t, float64(3), v.Value.(map[string]any)["x"], "the per-view override was restored")
		}
	}
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

// A Select whose unwired branch is taken carries a nil model, and the
// manifest drops those - which is how a part is turned off.
func TestSelectDropsAPartFromRenderAndExportWhenDisabled(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	body := makeSphereModel(t, session, "1", "0")
	whisker := makeSphereModel(t, session, "0.3", "2")

	var gate polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": logicSelectType,
		"inputs": map[string]any{
			"A":         map[string]any{"nodeId": whisker, "port": "Out"},
			"Condition": map[string]any{"value": "false"},
		},
	}, &gate)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{"Models": map[string]any{"elements": []any{
			map[string]any{"nodeId": body, "port": "Out"},
			map[string]any{"nodeId": gate.NodeId, "port": "Out"},
		}}},
	}, &manifest)

	dir := t.TempDir()
	var off polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{"nodeId": manifest.NodeId, "outputPath": filepath.Join(dir, "off.png")}, &off)

	callTool(t, session, "set_parameter", map[string]any{"nodeId": gate.NodeId, "port": "Condition", "value": "true"}, &polyformmcp.SetParameterOutput{})
	var on polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{"nodeId": manifest.NodeId, "outputPath": filepath.Join(dir, "on.png")}, &on)

	require.Greater(t, off.TriangleCount, 0, "the body still renders")
	assert.Greater(t, on.TriangleCount, off.TriangleCount, "enabling the gate adds the gated part's triangles")

	callTool(t, session, "set_parameter", map[string]any{"nodeId": gate.NodeId, "port": "Condition", "value": "false"}, &polyformmcp.SetParameterOutput{})
	callTool(t, session, "set_producer", map[string]any{"nodeId": manifest.NodeId, "port": "Out", "name": "gated.glb"}, &polyformmcp.SetProducerOutput{})
	var gen polyformmcp.GenerateOutput
	callTool(t, session, "generate", map[string]any{"outputDir": dir}, &gen)
	_, err := os.Stat(filepath.Join(dir, "gated.glb", "model.glb"))
	require.NoError(t, err, "a disabled gate exports cleanly as an empty node")
}

func TestDescribeManifestCountsEveryModelAndInstances(t *testing.T) {
	session, _ := testSessionWithInstance(t)
	body := makeSphereModel(t, session, "1", "0")
	child := makeSphereModel(t, session, "0.5", "2")

	var group polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfModelType,
		"inputs": map[string]any{
			"Name":     map[string]any{"value": `"Group"`},
			"Children": map[string]any{"elements": []any{map[string]any{"nodeId": child, "port": "Out"}}},
		},
	}, &group)

	var gate polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": logicSelectType,
		"inputs": map[string]any{
			"A":         map[string]any{"nodeId": body, "port": "Out"},
			"Condition": map[string]any{"value": "false"},
		},
	}, &gate)

	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{"Models": map[string]any{"elements": []any{
			map[string]any{"nodeId": body, "port": "Out"},
			map[string]any{"nodeId": group.NodeId, "port": "Out"},
			map[string]any{"nodeId": gate.NodeId, "port": "Out"},
		}}},
	}, &manifest)

	var out polyformmcp.DescribeManifestOutput
	callTool(t, session, "describe_manifest", map[string]any{"nodeId": manifest.NodeId}, &out)

	require.Len(t, out.Models, 3, "body, Group, Group/child - the disabled one is dropped, not listed empty: %+v", out.Models)
	assert.Equal(t, 1, out.EmptyModels, "only the group carries no mesh")
	sum := 0
	for _, m := range out.Models {
		sum += m.Triangles
	}
	assert.Equal(t, sum, out.TotalTriangles)
	assert.Greater(t, out.TotalTriangles, 0)

	var render polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{"nodeId": manifest.NodeId, "outputPath": filepath.Join(t.TempDir(), "m.png")}, &render)
	assert.Equal(t, render.TriangleCount, out.TotalTriangles, "the count matches what the renderer drew")
}
