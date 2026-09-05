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
	"github.com/EliCDavis/vector/vector3"
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
	Width      int                `json:"width,omitempty" jsonschema:"width in pixels of each view, defaults to 512. Every rendered pixel becomes context you carry for the rest of the build (roughly width*height/750 tokens per view, multiplied by the number of views), so leave this alone for routine checkpoints and raise it only to inspect a specific detail."`
	Height     int                `json:"height,omitempty" jsonschema:"height in pixels of each view, defaults to 384"`
	Scope      string             `json:"scope,omitempty"`
	Views      []RenderCameraView `json:"views,omitempty" jsonschema:"one or more camera views to render and composite into a single grid image at outputPath, instead of the one default angle. Use this to check several sides of a model in one call (e.g. front/back/top), or to get a precise, repeatable zoomed close-up on one part via target+zoom."`
	Exclude    []string           `json:"exclude,omitempty" jsonschema:"node ids to leave out of this render only - does not touch the graph. Use it to isolate which part causes a visual defect without disconnecting and reconnecting it. Each id must name a node with a glTF Model output; that model is skipped wherever it appears, nested under another model's Children as well as directly in the target's Models. Excludes a whole ModelNode's contribution, not an individual primitive inside its mesh-producing subgraph."`
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
	Mesh modeling.Mesh
	// Color is the material's BaseColorFactor, or a neutral gray stand-in
	// when the material sets none. ExplicitColor says which, because the
	// two are used differently: the stand-in may be shaded flat, but it
	// must not tint a mesh that carries its own vertex colors.
	Color         fauxgl.Color
	ExplicitColor bool
	Texture       fauxgl.Texture
}

func materialColor(mat *gltf.PolyformMaterial) (fauxgl.Color, bool) {
	if mat == nil || mat.PbrMetallicRoughness == nil || mat.PbrMetallicRoughness.BaseColorFactor == nil {
		return fauxgl.Color{R: 0.7, G: 0.7, B: 0.7, A: 1}, false
	}
	return fauxgl.MakeColor(mat.PbrMetallicRoughness.BaseColorFactor), true
}

// materialTexture returns the material's base color texture as a
// fauxgl.Texture, or nil if it has none. Only handles a texture already
// decoded into memory (PolyformTexture.Image, the common case for a
// procedurally-generated ColorTexture) - a texture referenced only by URI
// isn't resolved here, since that path isn't guaranteed to be a loadable
// file (could be a relative asset path, a data URI, etc.).
func materialTexture(mat *gltf.PolyformMaterial) fauxgl.Texture {
	if mat == nil || mat.PbrMetallicRoughness == nil || mat.PbrMetallicRoughness.BaseColorTexture == nil {
		return nil
	}
	img := mat.PbrMetallicRoughness.BaseColorTexture.Image
	if img == nil {
		return nil
	}
	return fauxgl.NewImageTexture(img)
}

// tintedVertexColor multiplies one vertex color by the material's base
// color factor, the way glTF combines COLOR_0 with baseColorFactor.
func tintedVertexColor(c vector3.Float64, tint fauxgl.Color) fauxgl.Color {
	return fauxgl.Color{
		R: c.X() * tint.R,
		G: c.Y() * tint.G,
		B: c.Z() * tint.B,
		A: 1,
	}
}

// drawGroup is one texture-homogeneous batch to rasterize in its own
// DrawMesh call - the depth buffer persists across calls on the same
// Context, so splitting a scene into several of these (one per part)
// still occludes correctly against each other, same as one combined draw.
type drawGroup struct {
	Mesh    *fauxgl.Mesh
	Texture fauxgl.Texture
}

// modelProducedBy returns the model value a node outputs, so an exclusion
// given as a node id can be matched anywhere in the scene tree.
func modelProducedBy(node nodes.Node) (*gltf.PolyformModel, error) {
	for _, port := range node.Outputs() {
		if out, ok := port.(nodes.Output[*gltf.PolyformModel]); ok {
			if model := out.Value(); model != nil {
				return model, nil
			}
			return nil, fmt.Errorf("it produces no model")
		}
	}
	return nil, fmt.Errorf("it has no glTF Model output, so it isn't a part that can be left out of a render")
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
func flattenModel(model *gltf.PolyformModel, parent trs.TRS, out *[]renderablePart, excluded map[*gltf.PolyformModel]bool) {
	if excluded[model] {
		return
	}

	local := trs.Identity()
	if model.TRS != nil {
		local = *model.TRS
	}

	if len(model.GpuInstances) > 0 {
		for _, instance := range model.GpuInstances {
			flattenModelAt(model, parent.Multiply(local.Multiply(instance)), out, excluded)
		}
		return
	}

	flattenModelAt(model, parent.Multiply(local), out, excluded)
}

// flattenModelAt bakes model's own mesh (and recurses into its children) at
// an already-fully-composed world transform - the per-instance body of
// flattenModel, factored out so a GpuInstances-bearing model can call it
// once per instance instead of once total.
func flattenModelAt(model *gltf.PolyformModel, world trs.TRS, out *[]renderablePart, excluded map[*gltf.PolyformModel]bool) {
	// An empty mesh has no Position attribute, and transforming one panics
	// deep in meshops with a message that names neither the model nor the
	// node that produced it. Skipping it here lets the caller report which
	// models were empty instead.
	if model.Mesh != nil && model.Mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		baked := model.Mesh.ApplyTRS(world)
		color, explicit := materialColor(model.Material)
		*out = append(*out, renderablePart{
			Mesh:          baked,
			Color:         color,
			ExplicitColor: explicit,
			Texture:       materialTexture(model.Material),
		})
	}

	for _, child := range model.Children {
		flattenModel(child, world, out, excluded)
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

// renderMeshView rasterizes one view of groups from eye looking at target.
// Each group is drawn in its own DrawMesh call against the same Context, so
// a per-part texture (or its absence) can be swapped in on the shared
// shader between draws while depth-testing still occludes correctly across
// every part, the same as if it had all been one combined mesh.
func renderMeshView(groups []drawGroup, width, height int, eye, target fauxgl.Vector, distance float64) image.Image {
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

	phong := fauxgl.NewPhongShader(matrix, dir, eye)
	phong.AmbientColor = fauxgl.Color{R: 0.45, G: 0.45, B: 0.45, A: 1}
	phong.DiffuseColor = fauxgl.Color{R: 0.65, G: 0.65, B: 0.65, A: 1}
	phong.SpecularColor = fauxgl.Color{R: 0.15, G: 0.15, B: 0.15, A: 1}

	dc := fauxgl.NewContext(width, height)
	dc.ClearColorBufferWith(fauxgl.Color{R: 0.85, G: 0.87, B: 0.9, A: 1})
	for _, g := range groups {
		phong.Texture = g.Texture
		dc.Shader = previewShader{phong}
		dc.DrawMesh(g.Mesh)
	}
	return dc.Image()
}

// previewShader is fauxgl's Phong shader with the specular term added
// rather than multiplied into the surface color.
//
// fauxgl returns color.Mul(ambient + diffuse + specular), which folds the
// highlight into the albedo. That is fine for mid-tone surfaces and
// destroys dark ones: a near-black body (albedo ~0.09, which is what a
// deep-sea creature or a black car actually is) can reach at most
// 0.09*1.25, about RGB 29, so every render came back a flat black
// silhouette with no readable form. A real glTF viewer shows that same
// model fine, because specular reflection off a dielectric is not tinted
// by albedo - which is why previews disagreed with the web viewer and
// sent builds into long blind debug loops.
//
// The vertex stage is inherited unchanged; only Fragment differs.
type previewShader struct {
	*fauxgl.PhongShader
}

func (s previewShader) Fragment(v fauxgl.Vertex) fauxgl.Color {
	color := v.Color
	if s.ObjectColor != fauxgl.Discard {
		color = s.ObjectColor
	}
	if s.Texture != nil {
		color = s.Texture.BilinearSample(v.Texture.X, v.Texture.Y)
	}

	diffuse := math.Max(v.Normal.Dot(s.LightDirection), 0)
	light := s.AmbientColor.Add(s.DiffuseColor.MulScalar(diffuse))
	if diffuse > 0 && s.SpecularPower > 0 {
		camera := s.CameraPosition.Sub(v.Position).Normalize()
		reflected := s.LightDirection.Negate().Reflect(v.Normal)
		if specular := math.Max(camera.Dot(reflected), 0); specular > 0 {
			specular = math.Pow(specular, s.SpecularPower)
			light = light.Add(s.SpecularColor.MulScalar(specular))
		}
	}
	shaded := color.Mul(light)

	return toneMap(shaded).Alpha(color.A)
}

// previewExposure and toneMap exist so a dark surface is legible.
//
// A preview is read by a model deciding whether the geometry is right, and
// a deep-sea creature or a black car is genuinely about 0.09 albedo. Under
// plain ambient+diffuse that spans 0.04 to 0.11 - a flat black silhouette
// with no readable form, whatever the lighting does, because the albedo
// multiplies everything. Real viewers show the same model fine because
// they tone map: highlights compress, so exposure can be raised far enough
// to lift shadow detail without blowing out the bright parts.
//
// Reinhard over a fixed exposure does that in one cheap step. It only
// touches model pixels; the background comes from the cleared buffer and
// keeps its exact color.
const previewExposure = 3.0

func toneMap(c fauxgl.Color) fauxgl.Color {
	// Compress luminance and carry the color along at the same ratio.
	// Tone mapping each channel on its own pulls whatever is brightest
	// toward white, which turned a saturated cyan lure pale.
	lum := 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
	if lum <= 0 {
		return fauxgl.Color{A: c.A}
	}
	exposed := lum * previewExposure
	scale := (exposed / (1 + exposed)) / lum

	clamp := func(v float64) float64 {
		if v > 1 {
			return 1
		}
		if v < 0 {
			return 0
		}
		return v
	}
	return fauxgl.Color{
		R: clamp(c.R * scale),
		G: clamp(c.G * scale),
		B: clamp(c.B * scale),
		A: c.A,
	}
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

		// Exclude by the model each node produces, not by the node, so a
		// ModelNode nested under another one's Children is skipped too.
		// Matching only top-level Models entries silently did nothing for
		// a nested part, and the two renders came back identical - which
		// reads as "this part isn't the problem" rather than "exclude
		// didn't apply", exactly backwards for the thing it is for.
		excluded := map[*gltf.PolyformModel]bool{}
		for _, id := range in.Exclude {
			node := inst.Node(id)
			if node == nil {
				return fmt.Errorf("can't exclude %q: no node exists with that id", id)
			}
			model, e := modelProducedBy(node)
			if e != nil {
				return fmt.Errorf("can't exclude %q: %w", id, e)
			}
			excluded[model] = true
		}

		var parts []renderablePart
		for _, port := range arrInput.Value() {
			modelOut, ok := port.(nodes.Output[*gltf.PolyformModel])
			if !ok || modelOut.Value() == nil {
				continue
			}
			flattenModel(modelOut.Value(), trs.Identity(), &parts, excluded)
		}
		if len(parts) == 0 {
			return fmt.Errorf("no renderable meshes found feeding node %q (excluding %d node(s)). Every model reaching it is empty - if any come from a March node, that usually means its Resolution is too low for its Domain (voxel size is 1/Resolution, and must be smaller than the domain) or the field never crosses the surface inside the domain", in.NodeId, len(in.Exclude))
		}

		var allTriangles []*fauxgl.Triangle // combined, for auto-framing only
		var groups []drawGroup
		for _, part := range parts {
			// A mesh with a per-vertex "Color" attribute (the standard
			// technique for SDF/marched meshes, which have no UVs to
			// texture) is shaded from that - PhongShader already
			// interpolates/lights whatever color is on each Vertex, so
			// this needs no shader changes, only reading the real
			// per-vertex value instead of repeating the material's flat
			// color on all three corners.
			//
			// glTF multiplies COLOR_0 by baseColorFactor rather than
			// letting either win, so tint does the same. Overriding
			// instead made a preview disagree with every real viewer: a
			// mesh carrying subtle near-white vertex variation over a
			// colored material rendered white here and correctly tinted
			// everywhere else.
			vertexColored := part.Mesh.HasFloat3Attribute(modeling.ColorAttribute)
			textured := part.Texture != nil && part.Mesh.HasFloat2Attribute(modeling.TexCoordAttribute)

			tint := fauxgl.Color{R: 1, G: 1, B: 1, A: 1}
			if part.ExplicitColor {
				tint = part.Color
			}

			var partTriangles []*fauxgl.Triangle
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
				switch {
				case textured:
					// PhongShader.Fragment samples the texture and ignores
					// vertex Color entirely whenever its Texture field is
					// set, so only the UV coordinates matter here.
					uv1 := tri.P1Vec2Attr(modeling.TexCoordAttribute)
					uv2 := tri.P2Vec2Attr(modeling.TexCoordAttribute)
					uv3 := tri.P3Vec2Attr(modeling.TexCoordAttribute)
					t.V1.Texture = fauxgl.V(uv1.X(), uv1.Y(), 0)
					t.V2.Texture = fauxgl.V(uv2.X(), uv2.Y(), 0)
					t.V3.Texture = fauxgl.V(uv3.X(), uv3.Y(), 0)
				case vertexColored:
					c1 := tri.P1Vec3Attr(modeling.ColorAttribute)
					c2 := tri.P2Vec3Attr(modeling.ColorAttribute)
					c3 := tri.P3Vec3Attr(modeling.ColorAttribute)
					t.V1.Color = tintedVertexColor(c1, tint)
					t.V2.Color = tintedVertexColor(c2, tint)
					t.V3.Color = tintedVertexColor(c3, tint)
				default:
					t.V1.Color = part.Color
					t.V2.Color = part.Color
					t.V3.Color = part.Color
				}
				partTriangles = append(partTriangles, t)
			}

			if len(partTriangles) == 0 {
				continue
			}
			allTriangles = append(allTriangles, partTriangles...)

			group := drawGroup{Mesh: fauxgl.NewTriangleMesh(partTriangles)}
			if textured {
				group.Texture = part.Texture
			}
			groups = append(groups, group)
		}

		// A preview is read by a model, and an image costs roughly
		// width*height/750 tokens of context for the rest of the build.
		// 512x384 is enough to judge silhouette, proportion and placement
		// - which is what a per-checkpoint render is for - at ~40% the
		// cost of 800x600. Inspecting a small detail is the caller's cue
		// to pass an explicit larger size, or a zoomed view, rather than
		// paying for the extra pixels on every routine render.
		width := in.Width
		if width <= 0 {
			width = 512
		}
		height := in.Height
		if height <= 0 {
			height = 384
		}

		box := fauxgl.NewTriangleMesh(allTriangles).BoundingBox()
		center := box.Center()
		size := box.Size()
		radius := math.Max(size.X, math.Max(size.Y, size.Z))
		if radius <= 0 {
			radius = 1
		}
		baseDistance := radius * 2.2

		if len(in.Views) == 0 {
			eye := center.Add(fauxgl.V(0.9, 0.7, 1.1).Normalize().MulScalar(baseDistance))
			img := renderMeshView(groups, width, height, eye, center, baseDistance)
			if e := writePNGWithDirs(in.OutputPath, img); e != nil {
				return e
			}

			out.Path = in.OutputPath
			out.Width = width
			out.Height = height
			out.TriangleCount = len(allTriangles)
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
			images[i] = renderMeshView(groups, width, height, eye, target, distance)
		}

		composite, cols, rows := compositeGrid(images, in.Views, width, height)
		if e := writePNGWithDirs(in.OutputPath, composite); e != nil {
			return e
		}

		out.Path = in.OutputPath
		out.Width = composite.Bounds().Dx()
		out.Height = composite.Bounds().Dy()
		out.TriangleCount = len(allTriangles)
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
		Description: "Rasterize a fast preview PNG of a gltf Manifest/Model scene with a CPU rasterizer. Surface color is the mesh's per-vertex \"Color\" attribute multiplied by the material's BaseColorFactor (as glTF combines COLOR_0 with baseColorFactor), or the flat factor alone with no vertex colors; a ColorTexture replaces both where the mesh has UVs, and is NOT multiplied by the factor here as a real viewer would, so keep a textured part's factor white. Output is tone mapped, so a near-black surface still shows readable form rather than a flat silhouette; it will look lighter than its raw color, which is what a real viewer shows too. With no views, auto-frames one default angle - use that for routine checkpoints. Pass views to composite several angles into one grid image, or target+zoom to inspect a detail. Pass exclude (node ids of Model-producing nodes) to drop parts from just this render, the non-destructive way to find which part causes a defect; it applies to nested Children as well as top-level Models, and an id that isn't a part is an error rather than a silently identical render.",
	}, s.renderPreview)
}
