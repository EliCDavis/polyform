package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"maps"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
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

func (t RenderTarget) vector() fauxgl.Vector { return fauxgl.V(t.X, t.Y, t.Z) }

func toRenderTarget(v fauxgl.Vector) RenderTarget { return RenderTarget{X: v.X, Y: v.Y, Z: v.Z} }

type RenderCameraView struct {
	Name         string            `json:"name,omitempty" jsonschema:"optional label captioned under this view when composited with others"`
	Azimuth      float64           `json:"azimuth,omitempty" jsonschema:"horizontal camera angle in degrees around the target: 0 = camera on +Z looking toward -Z (a canonical front view), 90 = camera on +X, 180 = camera on -Z, 270 (or -90) = camera on -X. Defaults to 0."`
	Elevation    float64           `json:"elevation,omitempty" jsonschema:"vertical camera angle in degrees: 0 = eye-level, 90 = straight down from above, -90 = straight up from below. Defaults to 0. Straight above or below, azimuth still turns the image: at azimuth 0, 90 puts -Z at the top of the image and -90 puts +Z there - the underside as it looks after flipping the model over its X axis."`
	Zoom         float64           `json:"zoom,omitempty" jsonschema:"distance multiplier on top of the auto-fit (or fitRadius): less than 1 moves closer, more than 1 backs out. Relative to whatever the fit chose, so it is not repeatable across scene changes - prefer frameOn or target+fitRadius for a close-up. Defaults to 1."`
	Target       *RenderTarget     `json:"target,omitempty" jsonschema:"world-space point to center this view on instead of the framed bounds' center. With fitRadius it is a predictable close-up of one spot; frameOn does the same for a whole part without computing the point by hand."`
	FitRadius    float64           `json:"fitRadius,omitempty" jsonschema:"frames a sphere of this radius (meters) around target whatever the scene's size: target at a joint with fitRadius 0.15 is a close-up of that joint, predictably. Replaces the auto-fit distance; zoom still multiplies it. Reported back in cameras[].fitRadius for every view, so the next call can reuse it."`
	Variables    map[string]string `json:"variables,omitempty" jsonschema:"variable overrides for this view only, on top of the call's 'variables' - an open / closed pose pair, or one look per view of a catalog sheet, in a single call. Framing stays that of the base scene so the views line up, unless frame is 'own' or the view frames on a node."`
	Profile      string            `json:"profile,omitempty" jsonschema:"a saved profile applied for this view only, under the view's own 'variables'; it captions the view when name is unset. One view per profile is a contact sheet of the saved looks."`
	Exclude      []string          `json:"exclude,omitempty" jsonschema:"parts left out of this view only, in addition to the call's 'exclude'; same ids as there."`
	Only         []string          `json:"only,omitempty" jsonschema:"render just these parts in this view, replacing the call's 'only'; same ids as 'exclude'."`
	Frame        string            `json:"frame,omitempty" jsonschema:"'own' fits this view to its own scene (after its variables and exclusions) instead of the base scene's bounds, so a look sheet whose looks differ in size - a tall variant beside a short one - shows each one whole. Default: shared base framing. frameOn and frameExclude imply 'own'."`
	FrameOn      []string          `json:"frameOn,omitempty" jsonschema:"fit this view to the world bounds of these parts only (node ids, instance ids, instance paths like outer/inner/node, or model names - what 'exclude' takes), GPU-instanced copies included; the rest of the scene is still drawn. The way to get a close-up of one part without computing its world position: frameOn its node and read the camera back from cameras[]."`
	FrameExclude []string          `json:"frameExclude,omitempty" jsonschema:"parts drawn but ignored when auto-fitting this view, in addition to the call's: a backdrop or ground slab that would otherwise shrink the subject."`
	Eye          *RenderTarget     `json:"eye,omitempty" jsonschema:"explicit camera position in world space; with it azimuth, elevation, zoom and fitRadius are ignored and the camera is used as given, never pulled clear of geometry. Pass a previous result's cameras[] eye/target/up/fov back to reproduce that exact framing."`
	LookAt       *RenderTarget     `json:"lookAt,omitempty" jsonschema:"what an explicit eye looks at; defaults to target, else the framed bounds' center."`
	Up           *RenderTarget     `json:"up,omitempty" jsonschema:"camera up for an explicit eye; defaults to world +Y (azimuth-turned when looking straight up or down)."`
	Fov          float64           `json:"fov,omitempty" jsonschema:"vertical field of view in degrees. Defaults to 28; the fit and fitRadius account for it."`
	Light        *RenderLight      `json:"light,omitempty" jsonschema:"key light for this view only, overriding the call's 'light'. Two views of one part under different light read shallow relief that neither shows alone."`
}

type RenderPreviewInput struct {
	NodeId     string `json:"nodeId" jsonschema:"id of a gltf.ManifestNode (or any node with a 'Models' array input) to rasterize - or, with 'port' or when the node has no Models input, a node whose glTF Model, Mesh or SDF field output is rendered on its own with a default gray material, no scaffolding needed. A field is marched for the render over bounds found by sampling it (see 'march' in the result); the graph is untouched."`
	Port       string `json:"port,omitempty" jsonschema:"output port of nodeId to render alone: a *gltf.PolyformModel, a modeling.Mesh, a sample.Vec3ToFloat SDF or an sdf.ColoredField. Omit for a manifest, or when the node has exactly one such output."`
	OutputPath string `json:"outputPath" jsonschema:"filesystem path to write the rendered PNG to"`
	Width      int    `json:"width,omitempty" jsonschema:"width in pixels of each view, defaults to 512. Every rendered pixel becomes context you carry for the rest of the build (roughly width*height/750 tokens per view, multiplied by the number of views), so leave this alone for routine checkpoints and raise it only to inspect a specific detail."`
	Height     int    `json:"height,omitempty" jsonschema:"height in pixels of each view, defaults to 384"`
	Scope      string `json:"scope,omitempty" jsonschema:"the root graph if omitted, or a subgraph id in which nodeId lives. On its own that is the subgraph DEFINITION, whose boundary inputs read as zero - pass instanceId to render a real instance. Never changes how 'exclude' and friends are looked up."`
	InstanceId string `json:"instanceId,omitempty" jsonschema:"id of a subgraph instance node in the root graph, or a path of instance ids (outer/inner) to one nested inside another part: nodeId is then a node inside that instance's own copy, with its wired inputs applied. The way to render one part of a scene in isolation."`
	Columns    int    `json:"columns,omitempty" jsonschema:"grid columns for 'views'. Default: as many as it takes to leave no empty cell (3 across for 3, 2x2 for 4, 5 across for 5)."`

	Views        []RenderCameraView `json:"views,omitempty" jsonschema:"one or more camera views to render and composite into a single grid image at outputPath, instead of the one default angle. Use this to check several sides of a model in one call (e.g. front/back/top), or for a close-up of one part via frameOn."`
	Exclude      []string           `json:"exclude,omitempty" jsonschema:"parts left out of this render only - the graph is untouched. Each entry is a node id, a subgraph instance id, an instance path (outer/inner/node, from the root, so 'scope' never changes it) or a model Name; every glTF model the named node's value reaches is dropped wherever it appears in the tree, GPU instances included. An entry that names nothing is an error rather than a silently identical render. Use it to isolate which part causes a defect without disconnecting and reconnecting it."`
	Only         []string           `json:"only,omitempty" jsonschema:"render only these parts (same ids as 'exclude'); everything else is left out of this render."`
	FrameOn      []string           `json:"frameOn,omitempty" jsonschema:"auto-fit every view to the world bounds of these parts (same ids as 'exclude') instead of the whole scene; the rest is still drawn. A view's own frameOn replaces it."`
	FrameExclude []string           `json:"frameExclude,omitempty" jsonschema:"parts drawn but ignored when auto-fitting (same ids as 'exclude'): a backdrop, a ground slab, a sky dome."`
	Variables    map[string]string  `json:"variables,omitempty" jsonschema:"variable path -> JSON value text, applied for this render only and restored afterward (a 'what would it look like with Base Color #9a9aa3 and Angle 0' check, or one look of a catalog sheet, without update_variable calls before and after). Values are JSON text as update_variable takes them."`
	Light        *RenderLight       `json:"light,omitempty" jsonschema:"where the key light sits relative to the camera, and how much fill. Omit for a three-quarter key that shows form. Raise 'azimuth' toward 0 to flatten shading, or swing it wide (60-80) with low 'ambient' to rake across shallow relief a front light cannot show."`

	ShowBackFaces bool `json:"showBackFaces,omitempty" jsonschema:"draw back faces solid magenta instead of culling them. A surface wound inside out is culled by viewers, so it looks missing or see-through; here it turns magenta, as does an open surface seen from behind."`

	Domain          *FieldProbeDomain `json:"domain,omitempty" jsonschema:"only when rendering an SDF field output: the box to march it in. Defaults to bounds found by sampling the field, which fails for an unbounded field (a ground plane)."`
	MarchResolution float64           `json:"marchResolution,omitempty" jsonschema:"only when rendering an SDF field output: cells per unit for the preview march, as MarchNode's Resolution. Defaults to 80 cells across the domain's longest side; the value used is reported in 'march'."`
}

type RenderCamera struct {
	View      string       `json:"view" jsonschema:"the view's name, or views[i]"`
	Eye       RenderTarget `json:"eye" jsonschema:"camera position in world space"`
	Target    RenderTarget `json:"target" jsonschema:"the point the camera looks at"`
	Up        RenderTarget `json:"up"`
	Fov       float64      `json:"fov" jsonschema:"vertical field of view in degrees"`
	Azimuth   float64      `json:"azimuth" jsonschema:"the eye's direction from the target, as a view's azimuth would give it"`
	Elevation float64      `json:"elevation"`
	Distance  float64      `json:"distance" jsonschema:"eye to target"`
	FitRadius float64      `json:"fitRadius" jsonschema:"the radius around target that fills the frame's height: pass it back with target to repeat this framing exactly, or scale it to tighten or loosen"`
	Near      float64      `json:"near" jsonschema:"near clip distance from the eye; geometry closer than this is cut away"`
	Far       float64      `json:"far"`
	Notes     []string     `json:"notes,omitempty" jsonschema:"anything done to or wrong with this view: a camera pulled in front of blocking geometry, a view that drew nothing"`
}

type RenderMarch struct {
	Domain     FieldProbeDomain `json:"domain" jsonschema:"the box the field was marched over"`
	Resolution float64          `json:"resolution" jsonschema:"cells per unit used, as MarchNode's Resolution"`
	CubeSize   float64          `json:"cubeSize"`
	Triangles  int              `json:"triangles"`
}

type RenderPreviewOutput struct {
	Path          string `json:"path"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	TriangleCount int    `json:"triangleCount"`
	Views         int    `json:"views,omitempty" jsonschema:"number of views composited into the image, if views was used"`
	Columns       int    `json:"columns,omitempty" jsonschema:"grid columns, if views was used"`
	Rows          int    `json:"rows,omitempty" jsonschema:"grid rows, if views was used"`

	Cameras  []RenderCamera `json:"cameras" jsonschema:"one per view (one for the default angle): the exact camera used, so the next call can reproduce or adjust it with eye/lookAt/up/fov or target/fitRadius"`
	Warnings []string       `json:"warnings,omitempty" jsonschema:"views that drew nothing but background, cameras moved to clear geometry - read these before trusting the image"`
	Rendered string         `json:"rendered,omitempty" jsonschema:"set when a single output was rendered in isolation rather than a manifest: which port, and how"`
	March    *RenderMarch   `json:"march,omitempty" jsonschema:"how an SDF field output was marched for this render"`

	EvaluationContext string `json:"evaluationContext,omitempty" jsonschema:"present when scope named a subgraph definition rather than a live instance; read it before trusting what was drawn"`
	EvalMs            int64  `json:"evalMs" jsonschema:"milliseconds spent evaluating the graph and gathering triangles"`
	RasterMs          int64  `json:"rasterMs" jsonschema:"milliseconds spent rasterizing, compositing and writing the PNG"`

	NodeErrors []NodeError `json:"nodeErrors,omitempty" jsonschema:"errors recorded by nodes that have run, anywhere in this graph including inside subgraph instances. A failed node does not stop the graph - it outputs an empty or zero value - so a part missing from the render is usually explained here"`
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
	Finish        finish
	Lineage       []*gltf.PolyformModel
}

// finish is a material's metallic and roughness. Unset is shaded as before
// they were read: the matte non-metal every earlier build was tuned against.
type finish struct {
	set                 bool
	metallic, roughness float64
}

func materialFinish(mat *gltf.PolyformMaterial) finish {
	if mat == nil || mat.PbrMetallicRoughness == nil {
		return finish{}
	}
	pbr := mat.PbrMetallicRoughness
	if pbr.MetallicFactor == nil && pbr.RoughnessFactor == nil {
		return finish{}
	}
	return finish{
		set:       true,
		metallic:  math.Min(math.Max(orDefault(pbr.MetallicFactor, 1), 0), 1),
		roughness: math.Min(math.Max(orDefault(pbr.RoughnessFactor, 1), 0), 1),
	}
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
	return srgbTexture{fauxgl.NewImageTexture(img)}
}

// glTF stores baseColorTexture sRGB-encoded while shading here runs in
// linear, so a texture sampled raw arrives already encoded and leaves
// encoded twice: a #8a5a30 oak renders as a washed-out #c4a078.
type srgbTexture struct{ fauxgl.Texture }

func (t srgbTexture) Sample(u, v float64) fauxgl.Color {
	return srgbToLinearColor(t.Texture.Sample(u, v))
}

func (t srgbTexture) BilinearSample(u, v float64) fauxgl.Color {
	return srgbToLinearColor(t.Texture.BilinearSample(u, v))
}

func srgbToLinearColor(c fauxgl.Color) fauxgl.Color {
	return fauxgl.Color{
		R: coloring.SRGBToLinear(c.R),
		G: coloring.SRGBToLinear(c.G),
		B: coloring.SRGBToLinear(c.B),
		A: c.A,
	}
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
	Finish  finish
}

// Composed transforms can shear, which no TRS holds, so positions go through
// the matrix itself. Normals only follow when the matrix decomposes cleanly;
// a sheared instance keeps its authored normals, which is fine for a preview.
func bakeTransform(mesh modeling.Mesh, world mat.Matrix4x4) modeling.Mesh {
	if t, err := trs.FromMatrix(world); err == nil {
		return mesh.ApplyTRS(t)
	}
	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	moved := make([]vector3.Float64, positions.Len())
	for i := range moved {
		moved[i] = world.MulPosition(positions.At(i))
	}
	return mesh.SetFloat3Attribute(modeling.PositionAttribute, moved)
}

// RenderLight aims the key light relative to the camera, so it keeps
// meaning the same thing when the view moves.
type RenderLight struct {
	Azimuth   *float64 `json:"azimuth,omitempty" jsonschema:"degrees the key light sits to the camera's right (negative for its left). 0 puts the light at the camera, which flattens shallow relief - a dome, a crease, a bevel - into no gradient at all. Defaults to -30."`
	Elevation *float64 `json:"elevation,omitempty" jsonschema:"degrees the key light sits above the camera (negative for below). Defaults to 25."`
	Ambient   *float64 `json:"ambient,omitempty" jsonschema:"fill light everywhere, 0 to 1. Lower deepens the shadow side and shows form more strongly; higher flattens it but keeps dark surfaces readable. Defaults to 0.45."`
}

const (
	defaultLightAzimuth   = -30.
	defaultLightElevation = 25.
	defaultLightAmbient   = 0.45
)

func orDefault(v *float64, fallback float64) float64 {
	if v == nil {
		return fallback
	}
	return *v
}

func (l RenderLight) ambient() float64 {
	return math.Min(math.Max(orDefault(l.Ambient, defaultLightAmbient), 0), 1)
}

// lightDirection swings toCamera by the light's angles in the camera's own
// basis. The result points from the surface toward the light, which is the
// direction fauxgl's Phong shader dots against the normal.
func lightDirection(toCamera, worldUp fauxgl.Vector, light RenderLight) fauxgl.Vector {
	right := worldUp.Cross(toCamera)
	if right.Length() < 1e-9 {
		right = fauxgl.V(1, 0, 0)
	}
	right = right.Normalize()
	up := toCamera.Cross(right).Normalize()

	az := orDefault(light.Azimuth, defaultLightAzimuth) * math.Pi / 180
	el := orDefault(light.Elevation, defaultLightElevation) * math.Pi / 180

	return toCamera.MulScalar(math.Cos(el) * math.Cos(az)).
		Add(right.MulScalar(math.Cos(el) * math.Sin(az))).
		Add(up.MulScalar(math.Sin(el))).
		Normalize()
}

var renderBackground = fauxgl.Color{R: 0.85, G: 0.87, B: 0.9, A: 1}

// renderMeshView rasterizes one view of groups from the pose. Each group is
// drawn in its own DrawMesh call against the same Context, so a per-part
// texture (or its absence) can be swapped in on the shared shader between
// draws while depth-testing still occludes correctly across every part, the
// same as if it had all been one combined mesh.
func renderMeshView(groups []drawGroup, width, height int, pose cameraPose, light RenderLight, backFaces bool) image.Image {
	dir := pose.direction()
	view := fauxgl.LookAt(pose.eye, pose.target, pose.up)
	proj := fauxgl.Perspective(pose.fov, float64(width)/float64(height), pose.near(), pose.far())
	// fauxgl accumulates its edge functions across each row, so an edge
	// running exactly through pixel centers is dropped by both triangles
	// that share it - a dotted crack across any centered quad. A fraction
	// of a pixel no axis-aligned geometry can land on keeps edges off them.
	offGrid := fauxgl.Translate(fauxgl.V(0.3137*2/float64(width), 0.2718*2/float64(height), 0))
	matrix := offGrid.Mul(proj).Mul(view)

	phong := fauxgl.NewPhongShader(matrix, lightDirection(dir, pose.up, light), pose.eye)
	ambient := light.ambient()
	phong.AmbientColor = fauxgl.Color{R: ambient, G: ambient, B: ambient, A: 1}
	phong.DiffuseColor = fauxgl.Color{R: 0.65, G: 0.65, B: 0.65, A: 1}
	phong.SpecularColor = fauxgl.Color{R: 0.15, G: 0.15, B: 0.15, A: 1}

	dc := fauxgl.NewContext(width, height)
	dc.ClearColorBufferWith(renderBackground)
	for _, g := range groups {
		phong.Texture = g.Texture
		dc.Shader = previewShader{phong, g.Finish}
		dc.DrawMesh(g.Mesh)
	}

	if backFaces {
		dc.Cull = fauxgl.CullFront
		dc.Shader = fauxgl.NewSolidColorShader(matrix, fauxgl.Color{R: 1, G: 0, B: 1, A: 1})
		for _, g := range groups {
			dc.DrawMesh(g.Mesh)
		}
	}
	return dc.Image()
}

// previewShader is fauxgl's Phong shader, tone mapped, with a material's
// metallic-roughness finish when it sets one. The vertex stage is inherited
// unchanged; only Fragment differs.
type previewShader struct {
	*fauxgl.PhongShader
	finish finish
}

func (s previewShader) Fragment(v fauxgl.Vertex) fauxgl.Color {
	color := v.Color
	if s.ObjectColor != fauxgl.Discard {
		color = s.ObjectColor
	}
	if s.Texture != nil {
		// glTF multiplies the texture by COLOR_0 rather than letting the
		// texture win, so dirt or wear painted per vertex still shows over
		// a textured part - which is what a real viewer draws.
		color = s.Texture.BilinearSample(v.Texture.X, v.Texture.Y).Mul(color)
	}
	if s.finish.set {
		return toneMap(s.finished(color, v)).Alpha(color.A)
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

// With no environment map a metal reflects nothing and renders near black, so
// a sky brightening upward stands in for one.
func (s previewShader) finished(albedo fauxgl.Color, v fauxgl.Vertex) fauxgl.Color {
	f := s.finish
	toCamera := s.CameraPosition.Sub(v.Position).Normalize()

	dielectric := fauxgl.Color{R: 0.04, G: 0.04, B: 0.04, A: 1}
	reflectance := dielectric.Lerp(albedo, f.metallic)

	lit := math.Max(v.Normal.Dot(s.LightDirection), 0)
	shaded := albedo.MulScalar(1 - f.metallic).Mul(s.AmbientColor.Add(s.DiffuseColor.MulScalar(lit)))

	// Capped so a glossy highlight still covers a few pixels of a preview.
	spread := math.Max(f.roughness*f.roughness, 0.01)
	if shininess := math.Min((2/(spread*spread)-2)/4, 128); lit > 0 && shininess > 0 {
		reflected := s.LightDirection.Negate().Reflect(v.Normal)
		highlight := math.Pow(math.Max(toCamera.Dot(reflected), 0), shininess) * (shininess + 2) / 8
		shaded = shaded.Add(reflectance.MulScalar(highlight * lit))
	}

	grazing := math.Pow(1-math.Max(v.Normal.Dot(toCamera), 0), 5) * (1 - f.roughness)
	rim := reflectance.Lerp(fauxgl.Color{R: 1, G: 1, B: 1, A: 1}, grazing)

	up := toCamera.Negate().Reflect(v.Normal).Y
	band := 0.05 + f.roughness
	sky := 0.15 + math.Min(math.Max((up+band)/(2*band), 0), 1)
	environment := sky + (0.65-sky)*f.roughness
	return shaded.Add(rim.MulScalar(environment))
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
//
// Material factors and vertex colors are linear in glTF, and every viewer
// applies the sRGB transfer curve on output. Skipping that curve here made a
// #c8322f collar preview as the saturated red of the hex while the exported
// GLB showed pale salmon - a whole session tuned colors against the wrong
// picture. The curve lifts midtones on its own, so the exposure that used to
// stand in for it comes down.
const previewExposure = 1.6

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

	encode := func(v float64) float64 {
		if v > 1 {
			v = 1
		}
		if v < 0 {
			v = 0
		}
		return coloring.LinearToSRGB(v)
	}
	return fauxgl.Color{
		R: encode(c.R * scale),
		G: encode(c.G * scale),
		B: encode(c.B * scale),
		A: c.A,
	}
}

// compositeGrid tiles rendered views into one image of cols columns, with
// an optional caption strip under each cell when any view has a Name. A
// spare cell is filled with the render background, not black, so it cannot
// be read as a failed view.
func compositeGrid(images []image.Image, views []RenderCameraView, cellW, cellH, cols int) (image.Image, int, int) {
	n := len(images)
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

	caption := func(x0, y0 int, label string, c color.Color) {
		d := &font.Drawer{
			Dst:  canvas,
			Src:  image.NewUniform(c),
			Face: basicfont.Face7x13,
			Dot: fixed.Point26_6{
				X: fixed.I(x0 + 4),
				Y: fixed.I(y0 + cellH + 15),
			},
		}
		d.DrawString(label)
	}

	for i := 0; i < cols*rows; i++ {
		col := i % cols
		row := i / cols
		x0 := col * cellW
		y0 := row * cellTotalH
		cell := image.Rect(x0, y0, x0+cellW, y0+cellH)
		if i >= n {
			draw.Draw(canvas, cell, &image.Uniform{C: renderBackground.NRGBA()}, image.Point{}, draw.Src)
			if hasCaptions {
				caption(x0, y0, "(unused)", color.Gray{Y: 140})
			}
			continue
		}
		draw.Draw(canvas, cell, images[i], image.Point{}, draw.Src)
		if hasCaptions {
			label := views[i].Name
			if label == "" {
				label = fmt.Sprintf("view %d", i+1)
			}
			caption(x0, y0, label, color.White)
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

// defaultViewDirection is the one angle a call with no views renders from.
var defaultViewDirection = fauxgl.V(0.9, 0.7, 1.1).Normalize()

func (s *Server) renderPreview(ctx context.Context, req *mcpsdk.CallToolRequest, in RenderPreviewInput) (*mcpsdk.CallToolResult, RenderPreviewOutput, error) {
	var out RenderPreviewOutput
	var err error
	s.atomic(&err, func() (e error) {
		inst, contextWarning, e := s.resolveFieldGraph(in.Scope, in.InstanceId)
		if e != nil {
			return e
		}
		out.EvaluationContext = contextWarning

		restore, e := s.overrideVariables(in.Variables)
		if e != nil {
			return e
		}
		defer func() { e = errors.Join(e, restore()) }()
		defer func() {
			out.NodeErrors = collectNodeErrors(inst)
			e = withNodeErrors(e, out.NodeErrors)
		}()

		var evalTime, rasterTime time.Duration
		timed := func(bucket *time.Duration, f func() error) error {
			start := time.Now()
			e := f()
			*bucket += time.Since(start)
			return e
		}

		callReq := sceneRequest{exclude: in.Exclude, only: in.Only, frameOn: in.FrameOn, frameExclude: in.FrameExclude}
		var base renderScene
		if e := timed(&evalTime, func() (e error) {
			base, e = s.collectScene(inst, in, callReq)
			return e
		}); e != nil {
			return e
		}
		out.Rendered, out.March = base.rendered, base.march
		out.Warnings = append(out.Warnings, base.warnings...)
		baseBox, _ := base.framingBox()

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
		aspect := float64(width) / float64(height)

		var baseLight RenderLight
		if in.Light != nil {
			baseLight = *in.Light
		}

		views := in.Views
		single := len(views) == 0
		if single {
			az, el := directionAngles(defaultViewDirection)
			views = []RenderCameraView{{Azimuth: az, Elevation: el}}
		}

		images := make([]image.Image, len(views))
		out.Cameras = make([]RenderCamera, len(views))
		for i, v := range views {
			label := fmt.Sprintf("views[%d]", i)
			if v.Name != "" {
				label = fmt.Sprintf("views[%d] (%s)", i, v.Name)
			}
			if v.FitRadius < 0 {
				return fmt.Errorf("%s: fitRadius must be positive", label)
			}
			if v.Fov < 0 || v.Fov >= 170 {
				return fmt.Errorf("%s: fov must be between 0 and 170 degrees", label)
			}
			switch v.Frame {
			case "", "base", "own":
			default:
				return fmt.Errorf("%s: frame must be 'own' or omitted, not %q", label, v.Frame)
			}

			overrides, e := s.viewOverrides(v)
			if e != nil {
				return fmt.Errorf("%s: %w", label, e)
			}
			if v.Name == "" {
				views[i].Name = v.Profile
			}

			// A view with its own variables or part lists re-collects the
			// scene (the refs resolve against the re-evaluated models) but
			// keeps the base framing unless it asks for its own.
			viewScene := base
			req := callReq
			if len(v.Exclude) > 0 {
				req.exclude = append(append([]string{}, in.Exclude...), v.Exclude...)
			}
			if len(v.Only) > 0 {
				req.only = v.Only
			}
			if len(v.FrameOn) > 0 {
				req.frameOn = v.FrameOn
			}
			if len(v.FrameExclude) > 0 {
				req.frameExclude = append(append([]string{}, in.FrameExclude...), v.FrameExclude...)
			}
			if len(overrides) > 0 || len(v.Exclude) > 0 || len(v.Only) > 0 || len(v.FrameOn) > 0 || len(v.FrameExclude) > 0 {
				if e := timed(&evalTime, func() error {
					viewRestore, e := s.overrideVariables(overrides)
					if e != nil {
						return e
					}
					viewScene, e = s.collectScene(inst, in, req)
					return errors.Join(e, viewRestore())
				}); e != nil {
					return fmt.Errorf("%s: %w", label, e)
				}
			}

			box := baseBox
			if v.Frame == "own" || len(v.FrameOn) > 0 || len(v.FrameExclude) > 0 {
				if own, ok := viewScene.framingBox(); ok {
					box = own
				}
			}

			fov := v.Fov
			if fov == 0 {
				fov = defaultFov
			}
			target := box.Center()
			if v.Target != nil {
				target = v.Target.vector()
			}
			dir := azimuthElevationDirection(v.Azimuth, v.Elevation)
			up := cameraUp(v.Azimuth, v.Elevation)

			var pose cameraPose
			var note string
			if v.Eye != nil {
				eye := v.Eye.vector()
				if v.LookAt != nil {
					target = v.LookAt.vector()
				}
				distance := eye.Sub(target).Length()
				if distance == 0 {
					return fmt.Errorf("%s: eye and lookAt are the same point", label)
				}
				dir = eye.Sub(target).DivScalar(distance)
				up = fauxgl.V(0, 1, 0)
				if math.Abs(dir.Y) > 0.999 {
					up = cameraUp(v.Azimuth, math.Copysign(90, dir.Y))
				}
				if v.Up != nil {
					if v.Up.vector().Length() == 0 {
						return fmt.Errorf("%s: up has zero length", label)
					}
					up = v.Up.vector().Normalize()
				}
				pose = cameraPose{eye: eye, target: target, up: up, distance: distance, fov: fov}
				if blockedAt, blocked := firstOccluder(viewScene.triangles, pose); blocked {
					note = fmt.Sprintf("geometry %.3g from lookAt stands between it and the explicit eye; the camera was left where you put it", blockedAt)
				}
			} else {
				distance := fitDistance(box, dir, up, aspect, fov)
				if v.FitRadius > 0 {
					distance = distanceForRadius(v.FitRadius, fov)
				}
				if v.Zoom > 0 {
					distance *= v.Zoom
				}
				pose = cameraPose{eye: target.Add(dir.MulScalar(distance)), target: target, up: up, distance: distance, fov: fov}
				// An inside-out part reads as a solid the camera sits in, and
				// showBackFaces is how that part gets found.
				if !in.ShowBackFaces {
					pose, note = clearCamera(pose, viewScene.triangles)
				}
			}

			viewLight := baseLight
			if v.Light != nil {
				viewLight = *v.Light
			}
			if e := timed(&rasterTime, func() error {
				images[i] = renderMeshView(viewScene.groups, width, height, pose, viewLight, in.ShowBackFaces)
				return nil
			}); e != nil {
				return e
			}

			az, el := directionAngles(pose.direction())
			camera := RenderCamera{
				View:      label,
				Eye:       toRenderTarget(pose.eye),
				Target:    toRenderTarget(pose.target),
				Up:        toRenderTarget(pose.up),
				Fov:       pose.fov,
				Azimuth:   az,
				Elevation: el,
				Distance:  pose.distance,
				FitRadius: fitRadiusOf(pose.distance, pose.fov),
				Near:      pose.near(),
				Far:       pose.far(),
			}
			if note != "" {
				camera.Notes = append(camera.Notes, note)
				out.Warnings = append(out.Warnings, label+": "+note)
			}
			if coverage(images[i], renderBackground.NRGBA()) == 0 {
				empty := "nothing but background was drawn: the framed region holds no geometry (target/fitRadius miss the subject, every part is excluded, or the eye is inside a solid)"
				camera.Notes = append(camera.Notes, empty)
				out.Warnings = append(out.Warnings, label+": "+empty)
			}
			out.Cameras[i] = camera
		}

		return timed(&rasterTime, func() error {
			defer func() {
				out.EvalMs = evalTime.Milliseconds()
				out.RasterMs = rasterTime.Milliseconds()
			}()
			out.Path = in.OutputPath
			out.TriangleCount = len(base.triangles)
			if single {
				out.Width, out.Height = width, height
				return writePNGWithDirs(in.OutputPath, images[0])
			}
			composite, cols, rows := compositeGrid(images, views, width, height, gridColumns(len(views), in.Columns))
			out.Width = composite.Bounds().Dx()
			out.Height = composite.Bounds().Dy()
			out.Views = len(views)
			out.Columns = cols
			out.Rows = rows
			return writePNGWithDirs(in.OutputPath, composite)
		})
	})
	return nil, out, err
}

// An unusable normal stays zero rather than borrowing the face's, so the
// preview shows the broken shading a real viewer would.
func vertexNormal(n vector3.Float64) fauxgl.Vector {
	if !usableNormal(n) {
		return fauxgl.Vector{}
	}
	return fauxgl.V(n.X(), n.Y(), n.Z()).Normalize()
}

func usableNormal(n vector3.Float64) bool {
	length := n.Length()
	return length > 1e-9 && !math.IsNaN(length) && !math.IsInf(length, 0)
}

func (s *Server) viewOverrides(v RenderCameraView) (map[string]string, error) {
	if v.Profile == "" {
		return v.Variables, nil
	}
	profile, err := s.graph.Profiles().Get(v.Profile)
	if err != nil {
		return nil, err
	}
	overrides := make(map[string]string, len(profile)+len(v.Variables))
	for path, value := range profile {
		overrides[path] = string(value)
	}
	maps.Copy(overrides, v.Variables)
	return overrides, nil
}

// overrideVariables applies path -> JSON value for the duration of one
// call; the returned func puts every variable back, in reverse order, so
// an override that fails part-way leaves nothing changed.
func (s *Server) overrideVariables(overrides map[string]string) (func() error, error) {
	type saved struct {
		path string
		data []byte
	}
	var snapshot []saved
	restore := func() error {
		var errs []error
		for i := len(snapshot) - 1; i >= 0; i-- {
			if _, e := s.graph.UpdateVariable(snapshot[i].path, snapshot[i].data); e != nil {
				errs = append(errs, fmt.Errorf("restoring variables[%q]: %w", snapshot[i].path, e))
			}
		}
		return errors.Join(errs...)
	}

	paths := make([]string, 0, len(overrides))
	for path := range overrides {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		value := overrides[path]
		if !json.Valid([]byte(value)) {
			return nil, errors.Join(fmt.Errorf("variables[%q] is not valid JSON: %s", path, value), restore())
		}
		before, e := s.graph.VariableData(path)
		if e != nil {
			return nil, errors.Join(fmt.Errorf("variables[%q]: %w", path, e), restore())
		}
		snapshot = append(snapshot, saved{path: path, data: before})
		if _, e := s.graph.UpdateVariable(path, []byte(value)); e != nil {
			return nil, errors.Join(fmt.Errorf("variables[%q]: %w", path, e), restore())
		}
	}
	return restore, nil
}

func (s *Server) registerRenderTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "render_preview",
		Description: "Rasterize a fast preview PNG of a gltf Manifest/Model scene with a CPU rasterizer, or of one node's Model, Mesh or SDF field output on its own (nodeId + port, with instanceId to reach inside a subgraph instance) - no throwaway manifest or material needed, and the graph is untouched. Surface color is the mesh's per-vertex \"Color\" attribute multiplied by the material's BaseColorFactor (as glTF combines COLOR_0 with baseColorFactor), or the flat factor alone with no vertex colors; a ColorTexture replaces both where the mesh has UVs, and is NOT multiplied by the factor here as a real viewer would, so keep a textured part's factor white. Colors are treated as glTF does - factors and vertex colors are linear multipliers, a ColorTexture is sRGB and is decoded on sample, and the result is sRGB-encoded on output - so the preview matches a real viewer: a raw hex pick like #c8322f renders pale as a factor (it is being read as linear), while the same hex baked into a texture renders as the hex; run screen-picked colors through coloring.SRGBToLinearNode (single colors) or meshops.SrgbToLinearNode (a vertex Color attribute) to get the color you see in the hex, and leave texture images alone. Output is also tone mapped, so a near-black surface still shows readable form rather than a flat silhouette. A material that sets MetallicFactor or RoughnessFactor is shaded with them - a metal reflects in its own color, a smooth surface takes a tight highlight - against a stand-in sky; one that sets neither is shaded as a matte non-metal, though glTF reads an unset MetallicFactor as fully metallic. Shading follows the mesh's Normal attribute where it has one, and is faceted where it has none, as in a viewer. With no views, auto-frames one default angle - use that for routine checkpoints. Pass views to composite several angles into one grid image laid out with no empty cells. Every view fits the camera to the projected bounds of what it frames - the whole scene, or the parts named by frameOn (a close-up of one part with no world point to compute), minus frameExclude - and the result's cameras[] gives back each view's eye, target, up, fov and fitRadius, so a framing can be repeated exactly (pass eye/lookAt/up/fov, or target/fitRadius) or tightened by scaling fitRadius. An auto-placed camera that lands behind or inside geometry is pulled in front of it and says so in cameras[].notes and warnings; a view that drew nothing but background is flagged the same way. Pass exclude (node ids, instance ids, instance paths or model names) to drop parts from just this render, and only to keep just the named ones - the non-destructive way to find which part causes a defect; an entry that names nothing is an error rather than a silently identical render. Pass variables (path -> JSON value) to render one look of a variant without touching the saved values: they are applied for this call and restored afterward, so a color or pose catalog is one call per look with no update_variable bookkeeping. evalMs and rasterMs say where the time went.",
	}, s.renderPreview)
}
