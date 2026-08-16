package mcp

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/fogleman/fauxgl"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type RenderTarget struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type RenderCameraView struct {
	Name      string        `json:"name,omitempty" jsonschema:"optional label captioned under this view when composited with others"`
	Azimuth   float64       `json:"azimuth,omitempty" jsonschema:"horizontal camera angle in degrees around the target: 0 = camera on +Z looking toward -Z (a canonical front view), 90 = camera on +X, 180 = camera on -Z, 270 (or -90) = camera on -X. Defaults to 0."`
	Elevation float64       `json:"elevation,omitempty" jsonschema:"vertical camera angle in degrees: 0 = eye-level, 90 = straight down from above, -90 = straight up from below. Defaults to 0."`
	Zoom      float64       `json:"zoom,omitempty" jsonschema:"distance multiplier from the default auto-framed distance; less than 1 moves the camera closer (e.g. 0.15 for a tight close-up), greater than 1 backs out. Must be positive. Defaults to 1."`
	Target    *RenderTarget `json:"target,omitempty" jsonschema:"world-space point to center this view on, overriding the whole scene's bounding-box center. Set this with a small Zoom to get a close-up of one specific part instead of framing the whole model."`
}

type RenderPreviewInput struct {
	NodeId     string             `json:"nodeId" jsonschema:"id of a gltf.ManifestNode (or any node with a 'Models' array input of *gltf.PolyformModel outputs) to rasterize"`
	OutputPath string             `json:"outputPath" jsonschema:"filesystem path to write the rendered PNG to"`
	Width      int                `json:"width,omitempty" jsonschema:"width in pixels of each view, defaults to 800"`
	Height     int                `json:"height,omitempty" jsonschema:"height in pixels of each view, defaults to 600"`
	Scope      string             `json:"scope,omitempty"`
	Views      []RenderCameraView `json:"views,omitempty" jsonschema:"one or more camera views to render and composite into a single grid image at outputPath, instead of the one default angle. Use this to check several sides of a model in one call (e.g. front/back/top), or to get a zoomed close-up on one part via target+zoom without hunting for it in a browser."`
	Exclude    []string           `json:"exclude,omitempty" jsonschema:"node ids to leave out of this render only - does not touch the graph, just skips these nodes' contribution when flattening the scene for this one call. Use this to isolate which part is responsible for a visual defect (e.g. render with a suspected part's node id excluded, compare) without disconnecting/reconnecting it in the actual graph. Matches entries in the target node's 'Models' array by the id of the node that produced each entry - a whole ModelNode's contribution is excluded, not an individual primitive nested inside its mesh-producing subgraph."`
}

type RenderPreviewOutput struct {
	Path          string `json:"path"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	TriangleCount int    `json:"triangleCount"`
	Views         int    `json:"views,omitempty" jsonschema:"number of views composited into the image, if views was used"`
	Columns       int    `json:"columns,omitempty" jsonschema:"grid columns, if views was used"`
	Rows          int    `json:"rows,omitempty" jsonschema:"grid rows, if views was used"`
}

type renderablePart struct {
	Mesh  modeling.Mesh
	Color fauxgl.Color
}

func materialColor(mat *gltf.PolyformMaterial) fauxgl.Color {
	if mat == nil || mat.PbrMetallicRoughness == nil || mat.PbrMetallicRoughness.BaseColorFactor == nil {
		return fauxgl.Color{R: 0.7, G: 0.7, B: 0.7, A: 1}
	}
	return fauxgl.MakeColor(mat.PbrMetallicRoughness.BaseColorFactor)
}

// flattenModel walks a PolyformModel tree, composing parent*local TRS at each
// level, and bakes the resulting world-space transform directly into a copy
// of each model's mesh so the caller can rasterize world-space triangles
// without tracking per-part matrices.
//
// A model with GpuInstances set is drawn once per instance instead of once
// at its own TRS - each instance's transform composes with the model's own
// TRS the same way the real glTF writer does
// (formats/gltf/model_trackers.go's instancesCachce.Add:
// model.TRS.Multiply(instance)), not as an additional single draw.
func flattenModel(model *gltf.PolyformModel, parent trs.TRS, out *[]renderablePart) {
	local := trs.Identity()
	if model.TRS != nil {
		local = *model.TRS
	}

	if len(model.GpuInstances) > 0 {
		for _, instance := range model.GpuInstances {
			flattenModelAt(model, parent.Multiply(local.Multiply(instance)), out)
		}
		return
	}

	flattenModelAt(model, parent.Multiply(local), out)
}

// flattenModelAt bakes model's own mesh (and recurses into its children) at
// an already-fully-composed world transform - the per-instance body of
// flattenModel, factored out so a GpuInstances-bearing model can call it
// once per instance instead of once total.
func flattenModelAt(model *gltf.PolyformModel, world trs.TRS, out *[]renderablePart) {
	if model.Mesh != nil {
		baked := model.Mesh.ApplyTRS(world)
		*out = append(*out, renderablePart{Mesh: baked, Color: materialColor(model.Material)})
	}

	for _, child := range model.Children {
		flattenModel(child, world, out)
	}
}

// azimuthElevationDirection returns the unit vector from a target toward a
// camera placed at the given horizontal/vertical angles (degrees).
func azimuthElevationDirection(azimuthDeg, elevationDeg float64) fauxgl.Vector {
	az := azimuthDeg * math.Pi / 180
	el := elevationDeg * math.Pi / 180
	return fauxgl.V(
		math.Cos(el)*math.Sin(az),
		math.Sin(el),
		math.Cos(el)*math.Cos(az),
	)
}

// renderMeshView rasterizes one view of mesh from eye looking at target.
func renderMeshView(mesh *fauxgl.Mesh, width, height int, eye, target fauxgl.Vector, distance float64) image.Image {
	up := fauxgl.V(0, 1, 0)
	dir := eye.Sub(target).Normalize()
	if math.Abs(dir.Y) > 0.999 {
		// eye is (near) directly above/below target - (0,1,0) would make
		// LookAt's cross product degenerate, so pick a different up axis.
		up = fauxgl.V(0, 0, -1)
	}

	view := fauxgl.LookAt(eye, target, up)
	proj := fauxgl.Perspective(28, float64(width)/float64(height), distance*0.05, distance*10)
	matrix := proj.Mul(view)

	shader := fauxgl.NewPhongShader(matrix, dir, eye)
	shader.AmbientColor = fauxgl.Color{R: 0.45, G: 0.45, B: 0.45, A: 1}
	shader.DiffuseColor = fauxgl.Color{R: 0.65, G: 0.65, B: 0.65, A: 1}
	shader.SpecularColor = fauxgl.Color{R: 0.15, G: 0.15, B: 0.15, A: 1}

	dc := fauxgl.NewContext(width, height)
	dc.ClearColorBufferWith(fauxgl.Color{R: 0.85, G: 0.87, B: 0.9, A: 1})
	dc.Shader = shader
	dc.DrawMesh(mesh)
	return dc.Image()
}

// compositeGrid tiles a list of rendered views into one image, columns x
// rows chosen to be as square as possible, with an optional caption strip
// under each cell when any view has a Name.
func compositeGrid(images []image.Image, views []RenderCameraView, cellW, cellH int) (image.Image, int, int) {
	n := len(images)
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := int(math.Ceil(float64(n) / float64(cols)))

	hasCaptions := false
	for _, v := range views {
		if v.Name != "" {
			hasCaptions = true
			break
		}
	}
	captionHeight := 0
	if hasCaptions {
		captionHeight = 20
	}
	cellTotalH := cellH + captionHeight

	canvas := image.NewNRGBA(image.Rect(0, 0, cols*cellW, rows*cellTotalH))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.NRGBA{R: 20, G: 20, B: 24, A: 255}}, image.Point{}, draw.Src)

	for i, img := range images {
		col := i % cols
		row := i / cols
		x0 := col * cellW
		y0 := row * cellTotalH
		draw.Draw(canvas, image.Rect(x0, y0, x0+cellW, y0+cellH), img, image.Point{}, draw.Src)

		if hasCaptions {
			label := views[i].Name
			if label == "" {
				label = fmt.Sprintf("view %d", i+1)
			}
			d := &font.Drawer{
				Dst:  canvas,
				Src:  image.NewUniform(color.White),
				Face: basicfont.Face7x13,
				Dot: fixed.Point26_6{
					X: fixed.I(x0 + 4),
					Y: fixed.I(y0 + cellH + 15),
				},
			}
			d.DrawString(label)
		}
	}

	return canvas, cols, rows
}

func writePNGWithDirs(path string, img image.Image) error {
	if dir := filepath.Dir(path); dir != "" {
		if e := os.MkdirAll(dir, 0o755); e != nil {
			return e
		}
	}
	return fauxgl.SavePNG(path, img)
}

func (s *Server) renderPreview(ctx context.Context, req *mcpsdk.CallToolRequest, in RenderPreviewInput) (*mcpsdk.CallToolResult, RenderPreviewOutput, error) {
	var out RenderPreviewOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}

		node := inst.Node(in.NodeId)
		if node == nil {
			return fmt.Errorf("no node exists with id %q", in.NodeId)
		}

		modelsInput, ok := node.Inputs()["Models"]
		if !ok {
			return fmt.Errorf("node %q has no \"Models\" input", in.NodeId)
		}
		arrInput, ok := modelsInput.(nodes.ArrayValueInputPort)
		if !ok {
			return fmt.Errorf("node %q's \"Models\" input is not an array port", in.NodeId)
		}

		excluded := map[nodes.Node]bool{}
		for _, id := range in.Exclude {
			excluded[inst.Node(id)] = true
		}

		var parts []renderablePart
		for _, port := range arrInput.Value() {
			modelOut, ok := port.(nodes.Output[*gltf.PolyformModel])
			if !ok || modelOut.Value() == nil || excluded[modelOut.Node()] {
				continue
			}
			flattenModel(modelOut.Value(), trs.Identity(), &parts)
		}
		if len(parts) == 0 {
			return fmt.Errorf("no renderable meshes found feeding node %q (excluding %d node(s))", in.NodeId, len(in.Exclude))
		}

		var triangles []*fauxgl.Triangle
		for _, part := range parts {
			// A mesh with a per-vertex "Color" attribute (the standard
			// technique for SDF/marched meshes, which have no UVs to
			// texture) is shaded from that instead of the flat material
			// color - PhongShader already interpolates/lights whatever
			// color is on each Vertex, so this needs no shader changes,
			// only reading the real per-vertex value instead of repeating
			// the material's flat color on all three corners.
			vertexColored := part.Mesh.HasFloat3Attribute(modeling.ColorAttribute)

			for i := 0; i < part.Mesh.PrimitiveCount(); i++ {
				tri := part.Mesh.Tri(i)
				p1 := tri.P1Vec3Attr(modeling.PositionAttribute)
				p2 := tri.P2Vec3Attr(modeling.PositionAttribute)
				p3 := tri.P3Vec3Attr(modeling.PositionAttribute)
				t := fauxgl.NewTriangleForPoints(
					fauxgl.V(p1.X(), p1.Y(), p1.Z()),
					fauxgl.V(p2.X(), p2.Y(), p2.Z()),
					fauxgl.V(p3.X(), p3.Y(), p3.Z()),
				)
				if vertexColored {
					c1 := tri.P1Vec3Attr(modeling.ColorAttribute)
					c2 := tri.P2Vec3Attr(modeling.ColorAttribute)
					c3 := tri.P3Vec3Attr(modeling.ColorAttribute)
					t.V1.Color = fauxgl.Color{R: c1.X(), G: c1.Y(), B: c1.Z(), A: 1}
					t.V2.Color = fauxgl.Color{R: c2.X(), G: c2.Y(), B: c2.Z(), A: 1}
					t.V3.Color = fauxgl.Color{R: c3.X(), G: c3.Y(), B: c3.Z(), A: 1}
				} else {
					t.V1.Color = part.Color
					t.V2.Color = part.Color
					t.V3.Color = part.Color
				}
				triangles = append(triangles, t)
			}
		}

		width := in.Width
		if width <= 0 {
			width = 800
		}
		height := in.Height
		if height <= 0 {
			height = 600
		}

		mesh := fauxgl.NewTriangleMesh(triangles)
		box := mesh.BoundingBox()
		center := box.Center()
		size := box.Size()
		radius := math.Max(size.X, math.Max(size.Y, size.Z))
		if radius <= 0 {
			radius = 1
		}
		baseDistance := radius * 2.2

		if len(in.Views) == 0 {
			eye := center.Add(fauxgl.V(0.9, 0.7, 1.1).Normalize().MulScalar(baseDistance))
			img := renderMeshView(mesh, width, height, eye, center, baseDistance)
			if e := writePNGWithDirs(in.OutputPath, img); e != nil {
				return e
			}

			out.Path = in.OutputPath
			out.Width = width
			out.Height = height
			out.TriangleCount = len(triangles)
			return nil
		}

		images := make([]image.Image, len(in.Views))
		for i, v := range in.Views {
			target := center
			if v.Target != nil {
				target = fauxgl.V(v.Target.X, v.Target.Y, v.Target.Z)
			}
			zoom := v.Zoom
			if zoom <= 0 {
				zoom = 1
			}
			distance := baseDistance * zoom
			eye := target.Add(azimuthElevationDirection(v.Azimuth, v.Elevation).MulScalar(distance))
			images[i] = renderMeshView(mesh, width, height, eye, target, distance)
		}

		composite, cols, rows := compositeGrid(images, in.Views, width, height)
		if e := writePNGWithDirs(in.OutputPath, composite); e != nil {
			return e
		}

		out.Path = in.OutputPath
		out.Width = composite.Bounds().Dx()
		out.Height = composite.Bounds().Dy()
		out.TriangleCount = len(triangles)
		out.Views = len(in.Views)
		out.Columns = cols
		out.Rows = rows
		return nil
	})
	return nil, out, err
}

func (s *Server) registerRenderTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "render_preview",
		Description: "Rasterize a fast preview PNG of a gltf Manifest/Model scene using a CPU software rasterizer. Reads a mesh's per-vertex \"Color\" attribute when present (the standard technique for SDF/marched meshes, which have no UVs to texture) - vertex-colored parts render with their real color here, not flat gray; a UV-mapped texture (glTF ColorTexture) still isn't sampled, only the material's flat BaseColorFactor is, for a part that has neither. With no views given, auto-frames a single default angle to the scene's bounding box (fastest, use this for the usual per-checkpoint sanity check). Pass views to render several camera angles and composite them into one grid image in a single call — e.g. front/back/top to catch a part hidden from the default angle, or one zoomed-in view (via target+zoom) to inspect a small detail without ever opening a browser. Pass exclude (a list of node ids) to leave specific parts out of just this one render, without touching the graph — the fast, non-destructive way to isolate which part is responsible for a visual defect, instead of disconnect/render/reconnect. Much faster than a ray-traced render or a browser round trip.",
	}, s.renderPreview)
}
