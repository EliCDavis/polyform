package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type DescribeManifestInput struct {
	NodeId            string            `json:"nodeId" jsonschema:"id of a node with a manifest output: a gltf.ManifestNode, or any node producing a manifest.Manifest"`
	Scope             string            `json:"scope,omitempty"`
	Variables         map[string]string `json:"variables,omitempty" jsonschema:"variable path -> JSON value text, applied for this call only and restored afterward, as render_preview's are: inspect one look of a variant without update_variable calls before and after"`
	Profile           string            `json:"profile,omitempty" jsonschema:"a saved profile applied for this call only, under 'variables'"`
	IncludeAttributes *bool             `json:"includeAttributes,omitempty" jsonschema:"whether each model's attributes carry their min/max. true always does, false never does, and omitting it does so for a manifest of 40 meshed models or fewer"`
	Models            string            `json:"models,omitempty" jsonschema:"full: every model with its world transform, bounds and attributes. compact: path, counts and material only. none: no per-model list. Omitted: full up to 40 models, compact past that"`
	Path              string            `json:"path,omitempty" jsonschema:"list only the models at or under this model path, e.g. 'Root/Part A'. Totals, materials, files and warnings still cover the whole file"`
}

const (
	manifestAttributeRangesInlineLimit = 40
	manifestFullModelsLimit            = 40
)

type ManifestAttributeSummary struct {
	Name      string    `json:"name" jsonschema:"the glTF semantic as the file names it: POSITION, NORMAL, COLOR_0, TEXCOORD_0, TANGENT (polyform's Color is COLOR_0, TexCoord is TEXCOORD_0)"`
	Type      string    `json:"type" jsonschema:"VEC2, VEC3 or VEC4"`
	Component string    `json:"component,omitempty" jsonschema:"float or ubyte"`
	Min       []float64 `json:"min,omitempty" jsonschema:"per component, as the accessor records it"`
	Max       []float64 `json:"max,omitempty"`
}

type ManifestModelSummary struct {
	Path            string `json:"path" jsonschema:"model names from the top-level entry down, joined with /"`
	Node            int    `json:"node" jsonschema:"index in the file's nodes array, which is what an animation channel's target names"`
	Triangles       int    `json:"triangles" jsonschema:"what a viewer draws for this model's own mesh, not counting children: the mesh once per GPU instance. 0 on a transform-only parent"`
	Vertices        int    `json:"vertices" jsonschema:"drawn, once per GPU instance like triangles"`
	Instances       int    `json:"instances,omitempty" jsonschema:"Gpu Instances count when the model is instanced"`
	StoredTriangles int    `json:"storedTriangles,omitempty" jsonschema:"the mesh as stored once in the file, when it is instanced"`
	StoredVertices  int    `json:"storedVertices,omitempty"`
	Empty           bool   `json:"empty,omitempty" jsonschema:"true when neither this model nor anything under it carries a mesh (a disabled gate, or a group whose parts all dropped out). A transform-only parent with meshed children is not empty"`
	Material        string `json:"material,omitempty" jsonschema:"the material it draws with, by its name under materials"`

	Attributes      []string                   `json:"attributes,omitempty" jsonschema:"every vertex attribute the file carries for this mesh, by glTF name"`
	AttributeRanges []ManifestAttributeSummary `json:"attributeRanges,omitempty" jsonschema:"each attribute's type and min/max; left out past 40 meshed models unless includeAttributes is set"`
	NotExported     []string                   `json:"notExported,omitempty" jsonschema:"mesh attributes the glTF writer drops: single-float ones such as Intensity or Opacity"`

	WorldPosition *Vector3Input `json:"worldPosition,omitempty" jsonschema:"where this model's origin lands after every parent transform - the frame a child part (one under a rotated parent) is actually placed in. Omitted at the origin"`
	WorldRotation *[4]float64   `json:"worldRotation,omitempty" jsonschema:"composed rotation as a quaternion (x, y, z, w). Omitted when unrotated"`
	WorldScale    *Vector3Input `json:"worldScale,omitempty" jsonschema:"omitted when 1, 1, 1"`
	Bounds        *ModelBounds  `json:"bounds,omitempty" jsonschema:"world-space box of this model's own mesh after every parent transform, covering every GPU instance when it has them; children carry their own"`
}

type ModelBounds struct {
	Min Vector3Input `json:"min"`
	Max Vector3Input `json:"max"`
}

func (m *ManifestModelSummary) compact() {
	m.WorldPosition, m.WorldRotation, m.WorldScale, m.Bounds = nil, nil, nil, nil
	m.Attributes, m.AttributeRanges, m.NotExported = nil, nil, nil
}

func (m *ManifestModelSummary) dropIdentity() {
	if p := m.WorldPosition; p != nil && *p == (Vector3Input{}) {
		m.WorldPosition = nil
	}
	if r := m.WorldRotation; r != nil && *r == [4]float64{0, 0, 0, 1} {
		m.WorldRotation = nil
	}
	if s := m.WorldScale; s != nil && *s == (Vector3Input{X: 1, Y: 1, Z: 1}) {
		m.WorldScale = nil
	}
}

type ManifestMaterialSummary struct {
	Name       string                    `json:"name" jsonschema:"the material's own name, or 'material N' by its index in the file when it has none"`
	Color      string                    `json:"color,omitempty" jsonschema:"baseColorFactor as hex. A linear multiplier, so a viewer shows it lighter than the hex reads."`
	Alpha      *float64                  `json:"alpha,omitempty" jsonschema:"baseColorFactor alpha when below 1"`
	AlphaMode  string                    `json:"alphaMode,omitempty"`
	Metallic   *float64                  `json:"metallic,omitempty" jsonschema:"omitted when unset, which a viewer reads as 1: fully metallic"`
	Roughness  *float64                  `json:"roughness,omitempty" jsonschema:"omitted when unset, which a viewer reads as 1"`
	Emissive   string                    `json:"emissive,omitempty" jsonschema:"emissiveFactor as hex"`
	Textures   []string                  `json:"textures,omitempty" jsonschema:"each texture slot it fills, with the image's size or the URI it points at"`
	Extensions map[string]map[string]any `json:"extensions,omitempty" jsonschema:"every extension written on this material, keyed by its glTF id (KHR_materials_transmission, ...), with the factors exactly as the file carries them; a texture slot reads 'texture N'"`
	Models     int                       `json:"models" jsonschema:"models drawn with it, instances counted once"`
}

type ManifestChannelSummary struct {
	Target        string  `json:"target" jsonschema:"path of the model the channel moves"`
	Node          int     `json:"node" jsonschema:"the node index the file points the channel at"`
	Path          string  `json:"path" jsonschema:"translation, rotation or scale"`
	Keys          int     `json:"keys"`
	Seconds       float64 `json:"seconds" jsonschema:"time of the last key"`
	Interpolation string  `json:"interpolation,omitempty"`
}

type ManifestAnimationSummary struct {
	Name     string                   `json:"name,omitempty"`
	Seconds  float64                  `json:"seconds" jsonschema:"the latest key over every channel"`
	Channels []ManifestChannelSummary `json:"channels"`
}

type ManifestFileSummary struct {
	Name          string `json:"name"`
	MeasuredBytes int64  `json:"measuredBytes" jsonschema:"exact size of the file generate writes, from encoding it once here; not an estimate"`
}

type DescribeManifestOutput struct {
	Models          []ManifestModelSummary `json:"models" jsonschema:"the models in the tree, depth first, as modelDetail and path say"`
	ModelDetail     string                 `json:"modelDetail" jsonschema:"how models was written: full, compact or none"`
	ModelCount      int                    `json:"modelCount" jsonschema:"models in the whole file, listed or not"`
	TotalTriangles  int                    `json:"totalTriangles" jsonschema:"what a viewer draws for this manifest, instances included"`
	TotalVertices   int                    `json:"totalVertices"`
	StoredTriangles int                    `json:"storedTriangles" jsonschema:"what the file stores: each shared or instanced mesh once"`
	StoredVertices  int                    `json:"storedVertices"`
	Bounds          *Extent                `json:"bounds,omitempty" jsonschema:"world-space box of everything drawn, instances included"`
	EmptyModels     int                    `json:"emptyModels" jsonschema:"models whose whole subtree has no mesh; a disabled gate is one, a transform-only parent of meshed children is not"`

	AttributeRangesOmitted bool `json:"attributeRangesOmitted,omitempty" jsonschema:"set when attribute min/max were left out for size; pass includeAttributes true to get them"`

	Materials          []ManifestMaterialSummary  `json:"materials,omitempty" jsonschema:"each material the file carries, in file order"`
	Animations         []ManifestAnimationSummary `json:"animations,omitempty"`
	ExtensionsUsed     []string                   `json:"extensionsUsed,omitempty" jsonschema:"the file's extensionsUsed, sorted"`
	ExtensionsRequired []string                   `json:"extensionsRequired,omitempty"`
	Warnings           []string                   `json:"warnings,omitempty" jsonschema:"material/mesh mismatches the file will carry silently: a normal texture or anisotropy on a mesh with no UVs or with zero-area UV triangles, one volume thickness shared by meshes of very different size, an animation aimed at a model that is not under Models"`
	Files              []ManifestFileSummary      `json:"files,omitempty" jsonschema:"each file the node writes and its size, as generate writes it"`

	NodeErrors []NodeError `json:"nodeErrors,omitempty" jsonschema:"errors recorded by nodes that ran for this description, anywhere in the graph including inside subgraph instances"`
}

func (s *Server) describeManifest(ctx context.Context, req *mcpsdk.CallToolRequest, in DescribeManifestInput) (*mcpsdk.CallToolResult, DescribeManifestOutput, error) {
	var out DescribeManifestOutput
	switch in.Models {
	case "", "full", "compact", "none":
	default:
		return nil, out, fmt.Errorf("models must be full, compact or none")
	}
	var err error
	s.atomic(&err, func() (e error) {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		node := lookupNode(inst, in.NodeId)
		if node == nil {
			return fmt.Errorf("no node exists with id %q", in.NodeId)
		}
		written, e := resolveManifestOutput(node, in.NodeId)
		if e != nil {
			return e
		}
		overrides, e := s.viewOverrides(RenderCameraView{Profile: in.Profile, Variables: in.Variables})
		if e != nil {
			return e
		}
		restore, e := s.overrideVariables(overrides)
		if e != nil {
			return e
		}
		defer func() { e = errors.Join(e, restore()) }()
		defer func() { out.NodeErrors = collectNodeErrors(inst) }()

		out.Models = []ManifestModelSummary{}
		entries := written.Value().Entries
		for _, name := range slices.Sorted(maps.Keys(entries)) {
			artifact := entries[name].Artifact
			if artifact == nil {
				continue
			}
			counted := &byteCount{}
			if glb, ok := artifact.(*gltf.Artifact); ok {
				e = describeGltfArtifact(glb, counted, &out)
			} else {
				e = artifact.Write(counted)
			}
			if e != nil {
				return fmt.Errorf("writing %s: %w", name, e)
			}
			out.Files = append(out.Files, ManifestFileSummary{Name: name, MeasuredBytes: counted.n})
		}

		meshed := 0
		for _, m := range out.Models {
			if m.Triangles > 0 {
				meshed++
			}
		}
		include := meshed <= manifestAttributeRangesInlineLimit
		if in.IncludeAttributes != nil {
			include = *in.IncludeAttributes
		}
		if !include {
			out.AttributeRangesOmitted = true
			for i := range out.Models {
				out.Models[i].AttributeRanges = nil
			}
		}
		if e := listModels(&out, in.Models, in.Path); e != nil {
			return e
		}
		out.round()
		return nil
	})
	return nil, out, err
}

func roundTo(v float64, decimals int) float64 {
	r, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', decimals, 64), 64)
	if r == 0 {
		return 0
	}
	return r
}

func roundAll(values []float64, decimals int) {
	for i, v := range values {
		values[i] = roundTo(v, decimals)
	}
}

func (e *Extent) round(decimals int) {
	if e == nil {
		return
	}
	for _, v := range []*Vector3Input{&e.Min, &e.Max, &e.Size, &e.Center} {
		v.round(decimals)
	}
}

func (v *Vector3Input) round(decimals int) {
	if v == nil {
		return
	}
	v.X, v.Y, v.Z = roundTo(v.X, decimals), roundTo(v.Y, decimals), roundTo(v.Z, decimals)
}

const manifestUnitDecimals = 4

func (out *DescribeManifestOutput) lengthDecimals() int {
	if out.Bounds == nil {
		return manifestUnitDecimals
	}
	size := max(out.Bounds.Size.X, out.Bounds.Size.Y, out.Bounds.Size.Z)
	if size <= 0 {
		return manifestUnitDecimals
	}
	return max(0, manifestUnitDecimals-int(math.Floor(math.Log10(size))))
}

func (out *DescribeManifestOutput) round() {
	lengths := out.lengthDecimals()
	out.Bounds.round(lengths)
	for i := range out.Models {
		m := &out.Models[i]
		m.WorldPosition.round(lengths)
		m.WorldScale.round(manifestUnitDecimals)
		if m.WorldRotation != nil {
			roundAll(m.WorldRotation[:], manifestUnitDecimals)
		}
		if m.Bounds != nil {
			m.Bounds.Min.round(lengths)
			m.Bounds.Max.round(lengths)
		}
		for a := range m.AttributeRanges {
			decimals := manifestUnitDecimals
			if m.AttributeRanges[a].Name == "POSITION" {
				decimals = lengths
			}
			roundAll(m.AttributeRanges[a].Min, decimals)
			roundAll(m.AttributeRanges[a].Max, decimals)
		}
		m.dropIdentity()
	}
}

func listModels(out *DescribeManifestOutput, detail, path string) error {
	out.ModelCount = len(out.Models)
	if path != "" {
		var under []ManifestModelSummary
		for _, m := range out.Models {
			if m.Path == path || strings.HasPrefix(m.Path, path+"/") {
				under = append(under, m)
			}
		}
		if len(under) == 0 {
			return fmt.Errorf("no model at path %q", path)
		}
		out.Models = under
	}

	out.ModelDetail = detail
	if detail == "" {
		out.ModelDetail = "full"
		if len(out.Models) > manifestFullModelsLimit {
			out.ModelDetail = "compact"
		}
	}
	switch out.ModelDetail {
	case "none":
		out.Models = []ManifestModelSummary{}
	case "compact":
		for i := range out.Models {
			out.Models[i].compact()
		}
	}
	return nil
}

func resolveManifestOutput(node nodes.Node, nodeID string) (nodes.Output[manifest.Manifest], error) {
	outputs := node.Outputs()
	for _, name := range sortedOutputKeys(outputs) {
		if written, ok := outputs[name].(nodes.Output[manifest.Manifest]); ok {
			return written, nil
		}
	}
	return nil, fmt.Errorf("node %q has no manifest output; it has %v", nodeID, sortedOutputNames(outputs))
}

type byteCount struct{ n int64 }

func (c *byteCount) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

// The writer encodes the scene once; the byte count and the structure both
// come from that one pass, so what is described is what generate writes.
func describeGltfArtifact(artifact *gltf.Artifact, counted io.Writer, out *DescribeManifestOutput) error {
	writer := gltf.NewWriter()
	writer.Options = artifact.Options
	if err := writer.AddScene(artifact.Scene); err != nil {
		return err
	}
	if err := writer.WriteGLB(counted, artifact.Options); err != nil {
		return err
	}

	walk := manifestWalk{
		out:             out,
		doc:             writer.ToGLTF(gltf.BufferEmbeddingStrategy_GLB),
		slots:           map[*gltf.PolyformModel]int{},
		storedIndices:   map[int]bool{},
		storedPositions: map[int]bool{},
		texturesListed:  map[int]bool{},
		volumeUses:      map[int][]volumeUse{},
	}
	walk.materials()
	for _, model := range artifact.Scene.Models {
		if model != nil {
			walk.model(model, "", trs.Identity())
		}
	}
	walk.volumeWarnings()
	walk.animations(artifact.Scene.Animations)
	out.ExtensionsUsed = mergeSorted(out.ExtensionsUsed, walk.doc.ExtensionsUsed)
	out.ExtensionsRequired = mergeSorted(out.ExtensionsRequired, walk.doc.ExtensionsRequired)
	return nil
}

type volumeUse struct {
	path      string
	across    float64
	thickness float64
}

type manifestWalk struct {
	out *DescribeManifestOutput
	doc gltf.Gltf

	// The writer appends a node after its children, so a post-order count
	// over the same tree is each model's index in doc.Nodes.
	next  int
	slots map[*gltf.PolyformModel]int

	storedIndices   map[int]bool
	storedPositions map[int]bool
	texturesListed  map[int]bool
	volumeUses      map[int][]volumeUse
	bounds          geometry.AABB
	hasBounds       bool
}

func (w *manifestWalk) warn(format string, args ...any) {
	w.out.Warnings = append(w.out.Warnings, fmt.Sprintf(format, args...))
}

func (w *manifestWalk) materials() {
	for i, m := range w.doc.Materials {
		summary := ManifestMaterialSummary{Name: m.Name}
		if summary.Name == "" {
			summary.Name = fmt.Sprintf("material %d", i+1)
		}
		if pbr := m.PbrMetallicRoughness; pbr != nil {
			if pbr.BaseColorFactor != nil {
				summary.Color = hexOf(pbr.BaseColorFactor[:3])
				if alpha := pbr.BaseColorFactor[3]; alpha < 1 {
					summary.Alpha = &alpha
				}
			}
			summary.Metallic, summary.Roughness = pbr.MetallicFactor, pbr.RoughnessFactor
		}
		if m.EmissiveFactor != nil {
			summary.Emissive = hexOf(m.EmissiveFactor[:])
		}
		if m.AlphaMode != nil {
			summary.AlphaMode = string(*m.AlphaMode)
		}
		for id, data := range m.Extensions {
			if summary.Extensions == nil {
				summary.Extensions = map[string]map[string]any{}
			}
			summary.Extensions[id] = plainExtension(data)
		}
		w.out.Materials = append(w.out.Materials, summary)
	}
}

func plainExtension(data any) map[string]any {
	fields, ok := data.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	plain := make(map[string]any, len(fields))
	for key, value := range fields {
		if tex, ok := value.(*gltf.TextureInfo); ok {
			plain[key] = fmt.Sprintf("texture %d", tex.Index)
			continue
		}
		plain[key] = value
	}
	return plain
}

func hexOf(rgb []float64) string {
	channel := func(v float64) int {
		return int(math.Round(math.Min(math.Max(v, 0), 1) * 255))
	}
	return fmt.Sprintf("#%02x%02x%02x", channel(rgb[0]), channel(rgb[1]), channel(rgb[2]))
}

func (w *manifestWalk) model(model *gltf.PolyformModel, parentPath string, parent trs.TRS) bool {
	path := model.Name
	if parentPath != "" {
		path = parentPath + "/" + model.Name
	}
	world := parent
	if model.TRS != nil {
		if composed, err := trs.FromMatrix(parent.Multiply(*model.TRS)); err == nil {
			world = composed
		}
	}
	rot := world.Rotation()
	position, scale := vector3Input(world.Position()), vector3Input(world.Scale())
	summary := ManifestModelSummary{
		Path:          path,
		WorldPosition: &position,
		WorldRotation: &[4]float64{rot.Dir().X(), rot.Dir().Y(), rot.Dir().Z(), rot.W()},
		WorldScale:    &scale,
	}
	slot := len(w.out.Models)
	w.out.Models = append(w.out.Models, summary)

	hasGeometry := false
	for _, child := range model.Children {
		if child != nil && w.model(child, path, world) {
			hasGeometry = true
		}
	}
	summary.Node = w.next
	w.next++
	w.slots[model] = slot

	if model.Mesh != nil && model.Mesh.PrimitiveCount() > 0 {
		hasGeometry = true
		w.geometry(model, world, &summary)
	} else if !hasGeometry {
		summary.Empty = true
		w.out.EmptyModels++
	}
	w.out.Models[slot] = summary
	return hasGeometry
}

func (w *manifestWalk) geometry(model *gltf.PolyformModel, world trs.TRS, summary *ManifestModelSummary) {
	mesh := *model.Mesh
	triangles := mesh.PrimitiveCount()
	vertices := mesh.AttributeLength()
	copies := max(1, len(model.GpuInstances))
	summary.Instances = len(model.GpuInstances)
	summary.Triangles = triangles * copies
	summary.Vertices = vertices * copies
	if copies > 1 {
		summary.StoredTriangles = triangles
		summary.StoredVertices = vertices
	}
	w.out.TotalTriangles += summary.Triangles
	w.out.TotalVertices += summary.Vertices

	local := mesh.BoundingBox(modeling.PositionAttribute)
	transforms := []mat.Matrix4x4{world.Matrix()}
	if len(model.GpuInstances) > 0 {
		transforms = transforms[:0]
		for _, instance := range model.GpuInstances {
			transforms = append(transforms, world.Multiply(instance))
		}
	}
	bounds := geometry.NewAABBFromPoints(worldCorners(local, transforms)...)
	summary.Bounds = &ModelBounds{Min: vector3Input(bounds.Min()), Max: vector3Input(bounds.Max())}
	if w.hasBounds {
		w.bounds.EncapsulateBounds(bounds)
	} else {
		w.bounds, w.hasBounds = bounds, true
	}
	w.out.Bounds = extentOf(w.bounds)

	summary.NotExported = mesh.Float1Attributes()
	sort.Strings(summary.NotExported)

	if summary.Node >= len(w.doc.Nodes) || w.doc.Nodes[summary.Node].Name != model.Name {
		return
	}
	docNode := w.doc.Nodes[summary.Node]
	if docNode.Mesh == nil || len(w.doc.Meshes[*docNode.Mesh].Primitives) == 0 {
		return
	}
	prim := w.doc.Meshes[*docNode.Mesh].Primitives[0]

	if prim.Indices != nil && !w.storedIndices[*prim.Indices] {
		w.storedIndices[*prim.Indices] = true
		w.out.StoredTriangles += triangles
	}
	if position, ok := prim.Attributes["POSITION"]; ok && !w.storedPositions[position] {
		w.storedPositions[position] = true
		w.out.StoredVertices += vertices
	}
	for _, name := range slices.Sorted(maps.Keys(prim.Attributes)) {
		accessor := w.doc.Accessors[prim.Attributes[name]]
		summary.Attributes = append(summary.Attributes, name)
		summary.AttributeRanges = append(summary.AttributeRanges, ManifestAttributeSummary{
			Name:      name,
			Type:      string(accessor.Type),
			Component: componentName(accessor.ComponentType),
			Min:       accessor.Min,
			Max:       accessor.Max,
		})
	}

	if prim.Material == nil {
		return
	}
	index := *prim.Material
	material := &w.out.Materials[index]
	material.Models++
	summary.Material = material.Name
	if !w.texturesListed[index] {
		w.texturesListed[index] = true
		material.Textures = textureSlots(model.Material)
	}
	w.materialMeshChecks(index, prim.Attributes, summary.Path, mesh, local)
}

func (w *manifestWalk) materialMeshChecks(index int, attributes map[string]gltf.GltfId, path string, mesh modeling.Mesh, local geometry.AABB) {
	docMaterial := w.doc.Materials[index]
	name := w.out.Materials[index].Name
	_, hasUV := attributes["TEXCOORD_0"]
	_, hasTangent := attributes["TANGENT"]
	_, anisotropic := docMaterial.Extensions["KHR_materials_anisotropy"]
	if docMaterial.NormalTexture != nil && !hasUV {
		w.warn("%s: material %q has a normal texture but the mesh has no TEXCOORD_0, so it cannot be sampled", path, name)
	}
	if anisotropic && !hasUV && !hasTangent {
		w.warn("%s: material %q uses KHR_materials_anisotropy but the mesh has no TEXCOORD_0 or TANGENT, so there is no tangent frame to orient it", path, name)
	}
	if (docMaterial.NormalTexture != nil || anisotropic) && hasUV && !hasTangent {
		if zero, total := zeroAreaUVTriangles(mesh); zero > 0 && zero*100 >= total {
			w.warn("%s: %d of %d triangles have zero-area UVs, so material %q's normal map or anisotropy has no direction there and viewers shade them wrong (often blown-out white)", path, zero, total, name)
		}
	}
	if volume, ok := docMaterial.Extensions["KHR_materials_volume"].(map[string]any); ok {
		if thickness, _ := volume["thicknessFactor"].(float64); thickness > 0 {
			size := local.Size()
			w.volumeUses[index] = append(w.volumeUses[index], volumeUse{
				path:      path,
				across:    max(size.X(), size.Y(), size.Z()),
				thickness: thickness,
			})
		}
	}
}

func zeroAreaUVTriangles(mesh modeling.Mesh) (zero, total int) {
	if mesh.Topology() != modeling.TriangleTopology || !mesh.HasFloat2Attribute(modeling.TexCoordAttribute) {
		return 0, 0
	}
	total = mesh.PrimitiveCount()
	for i := range total {
		tri := mesh.Tri(i)
		a := tri.P1Vec2Attr(modeling.TexCoordAttribute)
		ab := tri.P2Vec2Attr(modeling.TexCoordAttribute).Sub(a)
		ac := tri.P3Vec2Attr(modeling.TexCoordAttribute).Sub(a)
		if math.Abs(ab.X()*ac.Y()-ab.Y()*ac.X()) < 1e-12 {
			zero++
		}
	}
	return zero, total
}

func (w *manifestWalk) volumeWarnings() {
	for _, index := range slices.Sorted(maps.Keys(w.volumeUses)) {
		uses := w.volumeUses[index]
		small, big := uses[0], uses[0]
		for _, use := range uses[1:] {
			if use.across < small.across {
				small = use
			}
			if use.across > big.across {
				big = use
			}
		}
		if small.across > 0 && big.across > small.across*3 {
			w.warn("material %q: volume thicknessFactor %g is shared by %s (%.3g across) and %s (%.3g across); thickness is in mesh units, so it cannot suit both",
				w.out.Materials[index].Name, small.thickness, small.path, small.across, big.path, big.across)
		}
	}
}

func (w *manifestWalk) animations(animations []gltf.PolyformAnimation) {
	for i, animation := range animations {
		summary := ManifestAnimationSummary{Name: animation.Name, Channels: []ManifestChannelSummary{}}
		for c, channel := range animation.Channels {
			cs := ManifestChannelSummary{
				Path:          string(channel.TargetPath),
				Keys:          len(channel.Sampler.Times),
				Interpolation: string(channel.Sampler.Interpolation),
			}
			for _, t := range channel.Sampler.Times {
				cs.Seconds = max(cs.Seconds, t)
			}
			summary.Seconds = max(summary.Seconds, cs.Seconds)
			if i < len(w.doc.Animations) && c < len(w.doc.Animations[i].Channels) {
				if node := w.doc.Animations[i].Channels[c].Target.Node; node != nil {
					cs.Node = *node
				}
			}
			if slot, ok := w.slots[channel.Target]; ok {
				cs.Target = w.out.Models[slot].Path
			} else {
				name := "<nil>"
				if channel.Target != nil {
					name = channel.Target.Name
				}
				if cs.Node < len(w.doc.Nodes) {
					cs.Target = w.doc.Nodes[cs.Node].Name
				}
				w.warn("animation %q channel %d targets model %q, which is not under Models; the file aims it at node %d (%s) instead", animation.Name, c, name, cs.Node, cs.Target)
			}
			summary.Channels = append(summary.Channels, cs)
		}
		w.out.Animations = append(w.out.Animations, summary)
	}
}

func textureSlots(mat *gltf.PolyformMaterial) []string {
	if mat == nil {
		return nil
	}
	var slots []string
	slot := func(name string, tex *gltf.PolyformTexture) {
		switch {
		case tex == nil:
		case tex.Image != nil:
			size := tex.Image.Bounds().Size()
			slots = append(slots, fmt.Sprintf("%s %dx%d", name, size.X, size.Y))
		case tex.URI != "":
			slots = append(slots, fmt.Sprintf("%s %s", name, tex.URI))
		}
	}
	if pbr := mat.PbrMetallicRoughness; pbr != nil {
		slot("baseColor", pbr.BaseColorTexture)
		slot("metallicRoughness", pbr.MetallicRoughnessTexture)
	}
	if mat.NormalTexture != nil {
		slot("normal", mat.NormalTexture.PolyformTexture)
	}
	if mat.OcclusionTexture != nil {
		slot("occlusion", mat.OcclusionTexture.PolyformTexture)
	}
	slot("emissive", mat.EmissiveTexture)
	return slots
}

func componentName(t gltf.AccessorComponentType) string {
	switch t {
	case gltf.AccessorComponentType_FLOAT:
		return "float"
	case gltf.AccessorComponentType_UNSIGNED_BYTE:
		return "ubyte"
	case gltf.AccessorComponentType_UNSIGNED_SHORT:
		return "ushort"
	case gltf.AccessorComponentType_UNSIGNED_INT:
		return "uint"
	}
	return fmt.Sprint(int(t))
}

func worldCorners(local geometry.AABB, transforms []mat.Matrix4x4) []vector3.Float64 {
	lo, hi := local.Min(), local.Max()
	corners := make([]vector3.Float64, 0, 8*len(transforms))
	for _, m := range transforms {
		for _, x := range []float64{lo.X(), hi.X()} {
			for _, y := range []float64{lo.Y(), hi.Y()} {
				for _, z := range []float64{lo.Z(), hi.Z()} {
					corners = append(corners, m.MulPosition(vector3.New(x, y, z)))
				}
			}
		}
	}
	return corners
}

func extentOf(box geometry.AABB) *Extent {
	return &Extent{
		Min:    vector3Input(box.Min()),
		Max:    vector3Input(box.Max()),
		Size:   vector3Input(box.Size()),
		Center: vector3Input(box.Center()),
	}
}

func vector3Input(v vector3.Float64) Vector3Input {
	return Vector3Input{X: v.X(), Y: v.Y(), Z: v.Z()}
}

func mergeSorted(have, more []string) []string {
	for _, s := range more {
		if !slices.Contains(have, s) {
			have = append(have, s)
		}
	}
	sort.Strings(have)
	return have
}

func (s *Server) registerManifestTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "describe_manifest",
		Description: "What an export will contain, read from the glTF the writer builds, without writing or reading a file: per-model drawn and stored triangle/vertex counts, instance counts, world bounds (instances included), each model's glTF attributes (COLOR_0, TEXCOORD_0, NORMAL) with min/max, each material with every extension and its factors, the file's extensionsUsed/extensionsRequired, animations with per-channel targets, and the exact byte size of each file. This is the export check: never Read a .glb to verify it. A gated/disabled part shows as empty; a transform-only parent does not. Past 40 models the per-model list is compact (path, counts, material); pass path with models 'full' to see one part in detail. Takes variables and profile overrides like render_preview.",
	}, s.describeManifest)
}
