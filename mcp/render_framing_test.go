package mcp_test

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCubeModel(t *testing.T, session *mcpsdk.ClientSession, name string, w, h, d, x, z float64) string {
	t.Helper()
	var created polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType, "inputs": map[string]any{
			"Width": map[string]any{"value": num(w)}, "Height": map[string]any{"value": num(h)}, "Depth": map[string]any{"value": num(d)},
		}},
		map[string]any{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
			"Name":        map[string]any{"value": `"` + name + `"`},
			"Mesh":        map[string]any{"nodeId": "cube", "port": "Out"},
			"Translation": map[string]any{"value": `{"x":` + num(x) + `,"y":0,"z":` + num(z) + `}`},
		}},
	}}, &created)
	require.Empty(t, created.Errors)
	return created.Nodes["model"]
}

func num(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func manifestOf(t *testing.T, session *mcpsdk.ClientSession, models ...string) string {
	t.Helper()
	elements := make([]any, 0, len(models))
	for _, m := range models {
		elements = append(elements, map[string]any{"nodeId": m, "port": "Out"})
	}
	var manifest polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type":   gltfManifestType,
		"inputs": map[string]any{"Models": map[string]any{"elements": elements}},
	}, &manifest)
	return manifest.NodeId
}

func decodePNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	img, _, err := image.Decode(f)
	require.NoError(t, err)
	return img
}

func TestRenderPreviewRendersAMeshOutputAlone(t *testing.T) {
	session := testSession(t)
	var cube polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": cubeNodeType, "inputs": map[string]any{
		"Width": map[string]any{"value": "2"}, "Height": map[string]any{"value": "2"}, "Depth": map[string]any{"value": "2"},
	}}, &cube)

	path := filepath.Join(t.TempDir(), "mesh.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{"nodeId": cube.NodeId, "outputPath": path, "width": 64, "height": 64}, &out)

	assert.Equal(t, 12, out.TriangleCount)
	assert.Contains(t, out.Rendered, cube.NodeId+".Out")
	require.Len(t, out.Cameras, 1)
	assert.Greater(t, out.Cameras[0].Distance, 0.)
	assert.Empty(t, out.Warnings)

	img := decodePNG(t, path)
	assert.NotEqual(t, img.At(0, 0), img.At(32, 32), "the cube covers the center of the frame")
}

func TestRenderPreviewMarchesAFieldOutputOverItsOwnBounds(t *testing.T) {
	session := testSession(t)
	var sphere polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereNodeType, "inputs": map[string]any{
		"Radius":   map[string]any{"value": "0.5"},
		"Position": map[string]any{"value": `{"x":3,"y":0,"z":0}`},
	}}, &sphere)

	path := filepath.Join(t.TempDir(), "field.png")
	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{"nodeId": sphere.NodeId, "port": "Field", "outputPath": path, "width": 64, "height": 64}, &out)

	require.NotNil(t, out.March)
	assert.InDelta(t, 3, out.March.Domain.Center.X, 0.15, "the bounds follow the field, not the origin")
	assert.Greater(t, out.March.Domain.Size.X, 1.0)
	assert.Less(t, out.March.Domain.Size.X, 1.5, "the bounds hug the sphere")
	assert.Greater(t, out.March.Resolution, 0.)
	assert.Greater(t, out.TriangleCount, 100)
	assert.Contains(t, out.Rendered, "marched")

	img := decodePNG(t, path)
	assert.NotEqual(t, img.At(0, 0), img.At(32, 32), "the marched sphere fills the center of the frame")
}

func TestRenderPreviewRendersANodeInsideAnInstance(t *testing.T) {
	session := testSession(t)
	var sg polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "part", "name": "Part",
		"inputs": []map[string]any{{"name": "Radius", "type": "float64"}},
		"nodes": []map[string]any{
			{"alias": "sphere", "type": sdfSphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"nodeId": "Radius", "port": "Value"}}},
		},
		"outputs": []map[string]any{{"name": "Field", "nodeId": "sphere", "port": "Field"}},
	}, &sg)
	require.Empty(t, sg.Errors)
	var instance polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{
		"subgraphId": "part", "inputs": map[string]any{"Radius": map[string]any{"value": "2"}},
	}, &instance)

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": sg.Nodes["sphere"], "instanceId": instance.NodeId,
		"outputPath": filepath.Join(t.TempDir(), "inst.png"), "width": 32, "height": 32,
	}, &out)
	require.NotNil(t, out.March)
	assert.Greater(t, out.March.Domain.Size.X, 4.0, "the instance's wired radius, not the definition's zero, sets the bounds")
	assert.Empty(t, out.EvaluationContext)
}

func TestRenderPreviewRefusesAFieldItCannotBound(t *testing.T) {
	session := testSession(t)
	var plane polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": sdfSphereNodeType, "inputs": map[string]any{
		"Radius": map[string]any{"value": "1000"},
	}}, &plane)
	msg := callToolExpectingError(t, session, "render_preview", map[string]any{"nodeId": plane.NodeId, "outputPath": filepath.Join(t.TempDir(), "x.png")})
	assert.Contains(t, msg, "domain")
}

func TestRenderViewFramesOnANodeOrModelName(t *testing.T) {
	session := testSession(t)
	body := makeCubeModel(t, session, "Body", 10, 10, 10, 0, 0)
	knob := makeCubeModel(t, session, "Knob", 0.4, 0.4, 0.4, 8, 0)
	manifest := manifestOf(t, session, body, knob)

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(t.TempDir(), "frame.png"), "width": 64, "height": 64,
		"views": []map[string]any{
			{"azimuth": 0},
			{"azimuth": 0, "frameOn": []any{knob}},
			{"azimuth": 0, "frameOn": []any{"Knob"}},
		},
	}, &out)
	require.Len(t, out.Cameras, 3)

	whole, byId, byName := out.Cameras[0], out.Cameras[1], out.Cameras[2]
	assert.Greater(t, whole.FitRadius, 4.)
	assert.InDelta(t, 8, byId.Target.X, 1e-9, "the view centers on the knob")
	assert.Less(t, byId.FitRadius, 0.5)
	assert.Equal(t, byId.Eye, byName.Eye, "a model name frames the same part as its node id")
	assert.Equal(t, 3, out.Columns)
}

func TestRenderViewFrameExcludeIgnoresABackdropWhenFitting(t *testing.T) {
	session := testSession(t)
	body := makeCubeModel(t, session, "Body", 1, 1, 1, 0, 0)
	backdrop := makeCubeModel(t, session, "Backdrop", 20, 20, 0.1, 0, -3)
	manifest := manifestOf(t, session, body, backdrop)

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(t.TempDir(), "backdrop.png"), "width": 64, "height": 64,
		"views": []map[string]any{
			{"azimuth": 0},
			{"azimuth": 0, "frameExclude": []any{backdrop}},
		},
	}, &out)
	require.Len(t, out.Cameras, 2)
	assert.Greater(t, out.Cameras[0].FitRadius, 8.)
	assert.Less(t, out.Cameras[1].FitRadius, 1.5)
	assert.Greater(t, out.TriangleCount, 12, "the backdrop is still drawn")
}

func TestRenderCameraReportRoundTripsAsAnExplicitEye(t *testing.T) {
	session := testSession(t)
	manifest := manifestOf(t, session, makeCubeModel(t, session, "Body", 1, 2, 1.5, 0.3, 0))
	dir := t.TempDir()

	var first polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(dir, "a.png"), "width": 96, "height": 64,
		"views": []map[string]any{{"azimuth": 35, "elevation": 20}},
	}, &first)
	require.Len(t, first.Cameras, 1)
	cam := first.Cameras[0]
	assert.InDelta(t, 35, cam.Azimuth, 1e-6)
	assert.InDelta(t, 20, cam.Elevation, 1e-6)
	assert.InDelta(t, cam.Distance, math.Sqrt(math.Pow(cam.Eye.X-cam.Target.X, 2)+math.Pow(cam.Eye.Y-cam.Target.Y, 2)+math.Pow(cam.Eye.Z-cam.Target.Z, 2)), 1e-9)

	var second polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(dir, "b.png"), "width": 96, "height": 64,
		"views": []map[string]any{{
			"eye":    map[string]any{"x": cam.Eye.X, "y": cam.Eye.Y, "z": cam.Eye.Z},
			"lookAt": map[string]any{"x": cam.Target.X, "y": cam.Target.Y, "z": cam.Target.Z},
			"up":     map[string]any{"x": cam.Up.X, "y": cam.Up.Y, "z": cam.Up.Z},
			"fov":    cam.Fov,
		}},
	}, &second)
	assert.InDelta(t, cam.FitRadius, second.Cameras[0].FitRadius, 1e-9)

	a, b := decodePNG(t, filepath.Join(dir, "a.png")), decodePNG(t, filepath.Join(dir, "b.png"))
	differing := 0
	for y := 0; y < 64; y++ {
		for x := 0; x < 96; x++ {
			if a.At(x, y) != b.At(x, y) {
				differing++
			}
		}
	}
	assert.LessOrEqual(t, differing, 10, "the reported camera reproduces the picture")

	var third polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(dir, "c.png"), "width": 96, "height": 64,
		"views": []map[string]any{{
			"azimuth": 35, "elevation": 20,
			"target":    map[string]any{"x": cam.Target.X, "y": cam.Target.Y, "z": cam.Target.Z},
			"fitRadius": cam.FitRadius,
		}},
	}, &third)
	assert.InDelta(t, cam.Distance, third.Cameras[0].Distance, 1e-9, "target+fitRadius reproduce the same distance")
}

func TestRenderExcludeAndOnlyAcceptInstancesPathsAndModelNames(t *testing.T) {
	session := testSession(t)

	var gem polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "gem", "name": "Gem",
		"nodes": []map[string]any{
			{"alias": "sphere", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.5"}}},
			{"alias": "model", "type": gltfModelType, "inputs": map[string]any{
				"Name": map[string]any{"value": `"Gem"`},
				"Mesh": map[string]any{"nodeId": "sphere", "port": "Out"},
			}},
		},
		"outputs": []map[string]any{{"name": "Model", "nodeId": "model", "port": "Out"}},
	}, &gem)
	require.Empty(t, gem.Errors)
	var gemInstance polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "gem"}, &gemInstance)

	var cut polyformmcp.CreateSubgraphOutput
	callTool(t, session, "create_subgraph", map[string]any{
		"id": "cut", "name": "Cut",
		"nodes": []map[string]any{
			{"alias": "sphere", "type": sphereNodeType, "inputs": map[string]any{"Radius": map[string]any{"value": "0.3"}}},
		},
		"outputs": []map[string]any{{"name": "Mesh", "nodeId": "sphere", "port": "Out"}},
	}, &cut)
	require.Empty(t, cut.Errors)
	var cutInstance polyformmcp.InstantiateSubgraphOutput
	callTool(t, session, "instantiate_subgraph", map[string]any{"subgraphId": "cut"}, &cutInstance)
	var cutModel polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{"type": gltfModelType, "inputs": map[string]any{
		"Name":        map[string]any{"value": `"Facet"`},
		"Mesh":        map[string]any{"nodeId": cutInstance.NodeId, "port": "Mesh"},
		"Translation": map[string]any{"value": `{"x":-3,"y":0,"z":0}`},
	}}, &cutModel)

	body := makeCubeModel(t, session, "Body", 1, 1, 1, 3, 0)
	var manifestNode polyformmcp.CreateNodeOutput
	callTool(t, session, "create_node", map[string]any{
		"type": gltfManifestType,
		"inputs": map[string]any{"Models": map[string]any{"elements": []any{
			map[string]any{"nodeId": body, "port": "Out"},
			map[string]any{"nodeId": gemInstance.NodeId, "port": "Model"},
			map[string]any{"nodeId": cutModel.NodeId, "port": "Out"},
		}}},
	}, &manifestNode)
	manifest := manifestNode.NodeId
	dir := t.TempDir()

	render := func(args map[string]any) int {
		args["nodeId"] = manifest
		args["outputPath"] = filepath.Join(dir, "r.png")
		args["width"], args["height"] = 32, 32
		var out polyformmcp.RenderPreviewOutput
		callTool(t, session, "render_preview", args, &out)
		return out.TriangleCount
	}

	full := render(map[string]any{})
	withoutGem := render(map[string]any{"exclude": []any{gemInstance.NodeId}})
	require.Less(t, withoutGem, full, "an instance id drops the model it outputs")
	assert.Equal(t, withoutGem, render(map[string]any{"exclude": []any{gemInstance.NodeId + "/" + gem.Nodes["model"]}}), "an instance path reaches the node inside")
	assert.Equal(t, withoutGem, render(map[string]any{"exclude": []any{"Gem"}}), "a model name matches the same part")

	withoutFacet := render(map[string]any{"exclude": []any{cutInstance.NodeId}})
	assert.Less(t, withoutFacet, full, "an instance that outputs a mesh drops the model that mesh feeds")
	assert.Equal(t, full-withoutFacet, render(map[string]any{"only": []any{"Facet"}}), "only keeps just the named part")
	assert.Equal(t, 12, render(map[string]any{"only": []any{body}}))

	msg := callToolExpectingError(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(dir, "x.png"), "exclude": []any{"Nope"},
	})
	assert.Contains(t, msg, "no node exists")
}

func TestRenderGridLeavesNoEmptyCell(t *testing.T) {
	session := testSession(t)
	manifest := manifestOf(t, session, makeCubeModel(t, session, "Body", 1, 1, 1, 0, 0))
	dir := t.TempDir()

	views := func(n int) []map[string]any {
		out := make([]map[string]any, n)
		for i := range out {
			out[i] = map[string]any{"azimuth": float64(i) * 30}
		}
		return out
	}
	render := func(n int, args map[string]any) polyformmcp.RenderPreviewOutput {
		args["nodeId"] = manifest
		args["outputPath"] = filepath.Join(dir, "g.png")
		args["width"], args["height"] = 32, 32
		args["views"] = views(n)
		var out polyformmcp.RenderPreviewOutput
		callTool(t, session, "render_preview", args, &out)
		return out
	}

	three := render(3, map[string]any{})
	assert.Equal(t, 3, three.Columns)
	assert.Equal(t, 1, three.Rows)
	assert.Equal(t, 96, three.Width)

	five := render(5, map[string]any{})
	assert.Equal(t, 5, five.Columns)
	assert.Equal(t, 1, five.Rows)

	eight := render(8, map[string]any{})
	assert.Equal(t, 4, eight.Columns)
	assert.Equal(t, 2, eight.Rows)

	two := render(4, map[string]any{"columns": 2})
	assert.Equal(t, 2, two.Columns)
	assert.Equal(t, 2, two.Rows)

	seven := render(7, map[string]any{})
	assert.Equal(t, 3, seven.Columns)
	assert.Equal(t, 3, seven.Rows)
	img := decodePNG(t, filepath.Join(dir, "g.png"))
	assert.Equal(t, img.At(0, 0), img.At(32*2+16, 32*2+16), "a spare cell is filled with the render background, not black")
}

func TestRenderWarnsWhenAViewDrawsNothing(t *testing.T) {
	session := testSession(t)
	manifest := manifestOf(t, session, makeCubeModel(t, session, "Body", 1, 1, 1, 0, 0))

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(t.TempDir(), "empty.png"), "width": 32, "height": 32,
		"views": []map[string]any{
			{"name": "fine"},
			{"name": "missed", "target": map[string]any{"x": 1000, "y": 0, "z": 0}, "fitRadius": 0.1},
		},
	}, &out)
	require.Len(t, out.Warnings, 1)
	assert.Contains(t, out.Warnings[0], "missed")
	assert.Contains(t, out.Warnings[0], "nothing but background")
	assert.Empty(t, out.Cameras[0].Notes)
	assert.NotEmpty(t, out.Cameras[1].Notes)
}

func TestRenderCameraIsPulledInFrontOfBlockingGeometry(t *testing.T) {
	session := testSession(t)
	knob := makeCubeModel(t, session, "Knob", 0.2, 0.2, 0.2, 0, 0)
	wall := makeCubeModel(t, session, "Wall", 6, 6, 0.2, 0, 1)
	manifest := manifestOf(t, session, knob, wall)

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(t.TempDir(), "wall.png"), "width": 32, "height": 32,
		"views": []map[string]any{
			{"azimuth": 0, "target": map[string]any{"x": 0, "y": 0, "z": 0}, "fitRadius": 0.3},
			{"azimuth": 180, "target": map[string]any{"x": 0, "y": 0, "z": 0}, "fitRadius": 0.3},
		},
	}, &out)
	require.Len(t, out.Cameras, 2)

	blocked, clear := out.Cameras[0], out.Cameras[1]
	assert.Less(t, blocked.Eye.Z, 0.9, "the eye moved in front of the wall")
	assert.Greater(t, blocked.Eye.Z, 0.1, "but stayed outside the knob")
	assert.Greater(t, blocked.Fov, 28., "and widened to keep the framed radius")
	assert.InDelta(t, 0.3, blocked.FitRadius, 1e-9)
	require.NotEmpty(t, blocked.Notes)
	assert.Contains(t, blocked.Notes[0], "pulled in")
	assert.Len(t, out.Warnings, 1)

	assert.Empty(t, clear.Notes)
	assert.InDelta(t, -distanceFor(0.3), clear.Eye.Z, 1e-9)
	assert.Equal(t, 28., clear.Fov)
}

func distanceFor(radius float64) float64 {
	return 1.1 * radius / math.Tan(14*math.Pi/180)
}

func TestRenderTopDownViewFitsTheFootprintNotTheHeight(t *testing.T) {
	session := testSession(t)
	manifest := manifestOf(t, session, makeCubeModel(t, session, "Post", 0.5, 10, 0.5, 0, 0))

	var out polyformmcp.RenderPreviewOutput
	callTool(t, session, "render_preview", map[string]any{
		"nodeId": manifest, "outputPath": filepath.Join(t.TempDir(), "post.png"), "width": 32, "height": 32,
		"views": []map[string]any{{"azimuth": 0}, {"elevation": 90}},
	}, &out)
	require.Len(t, out.Cameras, 2)
	assert.Less(t, out.Cameras[1].Distance, out.Cameras[0].Distance/2, "from above a thin post is framed by its footprint")
	assert.Greater(t, out.Cameras[1].Distance, 5., "and the camera stays above its top")
}
