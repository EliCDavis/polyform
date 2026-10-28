package mcp

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/fauxgl"
)

// modelRefs is a set of glTF models named by a caller: by the node or
// instance that feeds them, or by their Name.
type modelRefs struct {
	models map[*gltf.PolyformModel]bool
	names  map[string]bool
}

func (r modelRefs) empty() bool {
	return len(r.models) == 0 && len(r.names) == 0
}

func (r modelRefs) matches(lineage []*gltf.PolyformModel) bool {
	for _, m := range lineage {
		if r.models[m] || r.names[m.Name] {
			return true
		}
	}
	return false
}

type sceneRequest struct {
	exclude, only, frameOn, frameExclude []string
}

type scenePart struct {
	triangles []*fauxgl.Triangle
	lineage   []*gltf.PolyformModel
}

type renderScene struct {
	groups    []drawGroup
	triangles []*fauxgl.Triangle
	parts     []scenePart

	frameOn, frameExclude modelRefs

	rendered string
	march    *RenderMarch
	warnings []string
}

func unusableNormals(m modeling.Mesh) (bad, total int) {
	if !m.HasFloat3Attribute(modeling.NormalAttribute) {
		return 0, 0
	}
	normals := m.Float3Attribute(modeling.NormalAttribute)
	for i := 0; i < normals.Len(); i++ {
		if !usableNormal(normals.At(i)) {
			bad++
		}
	}
	return bad, normals.Len()
}

func lineagePath(lineage []*gltf.PolyformModel) string {
	names := make([]string, len(lineage))
	for i, m := range lineage {
		names[i] = m.Name
	}
	return strings.Join(names, "/")
}

// framingBox is the bounds the camera fits: the frameOn parts, or every part,
// minus frameExclude. False when nothing is left to frame.
func (sc renderScene) framingBox() (fauxgl.Box, bool) {
	box := fauxgl.Box{}
	found := false
	for _, part := range sc.parts {
		if !sc.frameOn.empty() && !sc.frameOn.matches(part.lineage) {
			continue
		}
		if sc.frameExclude.matches(part.lineage) {
			continue
		}
		for _, tri := range part.triangles {
			for _, p := range []fauxgl.Vector{tri.V1.Position, tri.V2.Position, tri.V3.Position} {
				if !found {
					box, found = fauxgl.Box{Min: p, Max: p}, true
					continue
				}
				box.Min, box.Max = box.Min.Min(p), box.Max.Max(p)
			}
		}
	}
	return box, found
}

func (s *Server) collectScene(inst *graph.Graph, in RenderPreviewInput, req sceneRequest) (renderScene, error) {
	node := lookupNode(inst, in.NodeId)
	if node == nil {
		return renderScene{}, s.explainMissingNode(in.Scope, in.NodeId, fmt.Errorf("no node exists with id %q", in.NodeId))
	}
	subject, err := renderSubjectOf(node, in)
	if err != nil {
		return renderScene{}, err
	}

	exclude, err := s.resolveModelRefs(inst, req.exclude, "exclude")
	if err != nil {
		return renderScene{}, err
	}
	only, err := s.resolveModelRefs(inst, req.only, "render only")
	if err != nil {
		return renderScene{}, err
	}
	frameOn, err := s.resolveModelRefs(inst, req.frameOn, "frame on")
	if err != nil {
		return renderScene{}, err
	}
	frameExclude, err := s.resolveModelRefs(inst, req.frameExclude, "leave out of framing")
	if err != nil {
		return renderScene{}, err
	}

	scene, err := buildScene(subject.models, exclude, only, in.NodeId, len(req.exclude))
	if err != nil {
		return renderScene{}, err
	}
	for verb, refs := range map[string]modelRefs{"exclude": exclude, "render only": only, "frame on": frameOn, "leave out of framing": frameExclude} {
		if e := unmatchedNames(refs, scene.parts, verb); e != nil {
			return renderScene{}, e
		}
	}
	scene.frameOn, scene.frameExclude = frameOn, frameExclude
	scene.rendered, scene.march = subject.rendered, subject.march
	return scene, nil
}

func unmatchedNames(refs modelRefs, parts []scenePart, verb string) error {
	names := make([]string, 0, len(refs.names))
	for name := range refs.names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		matched := false
		for _, part := range parts {
			for _, m := range part.lineage {
				if m.Name == name {
					matched = true
				}
			}
		}
		if !matched {
			return fmt.Errorf("can't %s %q: no node exists with that id and no model in the scene is named that", verb, name)
		}
	}
	return nil
}

// resolveModelRefs reads each id as a node id, an instance path
// (outer/inner/node) from the root, or - when neither exists - a model Name
// to be matched while the scene is walked. Ids are looked up from the root
// first, so a caller's scope never changes what an id means.
func (s *Server) resolveModelRefs(inst *graph.Graph, ids []string, verb string) (modelRefs, error) {
	refs := modelRefs{models: map[*gltf.PolyformModel]bool{}, names: map[string]bool{}}
	for _, id := range ids {
		models, found, err := s.modelsFedBy(inst, id)
		if err != nil {
			return refs, fmt.Errorf("can't %s %q: %w", verb, id, err)
		}
		if !found {
			refs.names[id] = true
			continue
		}
		if len(models) == 0 {
			return refs, fmt.Errorf("can't %s %q: nothing it outputs reaches a glTF Model (it produces no model, or a gate ahead of the model is off)", verb, id)
		}
		for _, m := range models {
			refs.models[m] = true
		}
	}
	return refs, nil
}

type instanceHop struct {
	parent   *graph.Graph
	instance *graph.SubgraphInstanceNode
}

// modelsFedBy follows a node's value downstream to the glTF models it ends
// up in, stopping at the first model-producing node so a part never drags
// the group it sits in along with it. A value lifted out of an instance is
// followed on from the instance node in the parent graph.
func (s *Server) modelsFedBy(inst *graph.Graph, id string) ([]*gltf.PolyformModel, bool, error) {
	segments := strings.Split(id, "/")
	g := s.graph.Graph
	var hops []instanceHop
	for _, seg := range segments[:len(segments)-1] {
		node := lookupNode(g, seg)
		if node == nil {
			return nil, false, nil
		}
		sub, ok := node.(*graph.SubgraphInstanceNode)
		if !ok {
			return nil, false, fmt.Errorf("%q is not a subgraph instance, so there is nothing inside it to name", seg)
		}
		live := sub.LiveGraph()
		if live == nil {
			return nil, false, fmt.Errorf("instance %q has no graph to look inside", seg)
		}
		hops = append(hops, instanceHop{g, sub})
		g = live
	}
	last := segments[len(segments)-1]
	node := lookupNode(g, last)
	if node == nil && len(segments) == 1 && inst != s.graph.Graph {
		g, node = inst, lookupNode(inst, last)
	}
	if node == nil {
		return nil, false, nil
	}

	var models []*gltf.PolyformModel
	seen := map[nodes.Node]bool{}
	var visit func(g *graph.Graph, n nodes.Node, hops []instanceHop)
	visit = func(g *graph.Graph, n nodes.Node, hops []instanceHop) {
		if seen[n] {
			return
		}
		seen[n] = true
		if produced := modelsOutputBy(n); len(produced) > 0 {
			models = append(models, produced...)
			return
		}
		if _, lifted := n.(*subgraph.OutputNode); lifted && len(hops) > 0 {
			up := hops[len(hops)-1]
			visit(up.parent, up.instance, hops[:len(hops)-1])
			return
		}
		for _, consumer := range consumersOf(g, n) {
			visit(g, consumer, hops)
		}
	}
	visit(g, node, hops)
	return models, true, nil
}

func modelsOutputBy(n nodes.Node) []*gltf.PolyformModel {
	outputs := n.Outputs()
	var models []*gltf.PolyformModel
	for _, name := range sortedOutputKeys(outputs) {
		switch out := outputs[name].(type) {
		case nodes.Output[*gltf.PolyformModel]:
			if m := out.Value(); m != nil {
				models = append(models, m)
			}
		case nodes.Output[[]*gltf.PolyformModel]:
			for _, m := range out.Value() {
				if m != nil {
					models = append(models, m)
				}
			}
		}
	}
	return models
}

func consumersOf(g *graph.Graph, producer nodes.Node) []nodes.Node {
	ids := g.NodeIds()
	sort.Strings(ids)
	var consumers []nodes.Node
	for _, id := range ids {
		candidate := g.Node(id)
		for _, input := range candidate.Inputs() {
			if fedBy(input, producer) {
				consumers = append(consumers, candidate)
				break
			}
		}
	}
	return consumers
}

func fedBy(input nodes.InputPort, producer nodes.Node) bool {
	switch port := input.(type) {
	case nodes.SingleValueInputPort:
		src := port.Value()
		return src != nil && src.Node() == producer
	case nodes.ArrayValueInputPort:
		for _, src := range port.Value() {
			if src != nil && src.Node() == producer {
				return true
			}
		}
	}
	return false
}

type renderSubject struct {
	models   []*gltf.PolyformModel
	rendered string
	march    *RenderMarch
}

// renderSubjectOf is what the call draws: the models feeding a manifest, or
// one named output rendered on its own.
func renderSubjectOf(node nodes.Node, in RenderPreviewInput) (renderSubject, error) {
	outputs := node.Outputs()
	if in.Port == "" {
		if arr, ok := node.Inputs()["Models"].(nodes.ArrayValueInputPort); ok {
			var models []*gltf.PolyformModel
			for _, port := range arr.Value() {
				if out, ok := port.(nodes.Output[*gltf.PolyformModel]); ok && out.Value() != nil {
					models = append(models, out.Value())
				}
			}
			return renderSubject{models: models}, nil
		}
	}

	name := in.Port
	if name == "" {
		candidates := renderableOutputs(outputs)
		switch len(candidates) {
		case 0:
			return renderSubject{}, fmt.Errorf("node %q has no \"Models\" input and no glTF Model, Mesh or SDF field output to render; its outputs are %s", in.NodeId, strings.Join(sortedOutputNames(outputs), ", "))
		case 1:
			name = candidates[0]
		default:
			labels := make([]string, len(candidates))
			for i, key := range candidates {
				labels[i] = portLabel(outputs[key])
			}
			return renderSubject{}, fmt.Errorf("node %q has several renderable outputs (%s); say which with 'port'", in.NodeId, strings.Join(labels, ", "))
		}
	} else {
		name = resolveOutputPortName(node, name)
		if _, ok := outputs[name]; !ok {
			return renderSubject{}, fmt.Errorf("node %q has no output port %q; it has %s", in.NodeId, in.Port, strings.Join(sortedOutputNames(outputs), ", "))
		}
	}

	label := in.NodeId + "." + name
	switch out := outputs[name].(type) {
	case nodes.Output[*gltf.PolyformModel]:
		m := out.Value()
		if m == nil {
			return renderSubject{}, fmt.Errorf("%s is a nil model (a gate ahead of it is off)", label)
		}
		return renderSubject{models: []*gltf.PolyformModel{m}, rendered: label + " (glTF model, with its own materials)"}, nil
	case nodes.Output[[]*gltf.PolyformModel]:
		var models []*gltf.PolyformModel
		for _, m := range out.Value() {
			if m != nil {
				models = append(models, m)
			}
		}
		return renderSubject{models: models, rendered: fmt.Sprintf("%s (%d glTF models, with their own materials)", label, len(models))}, nil
	case nodes.Output[modeling.Mesh]:
		mesh := out.Value()
		return renderSubject{models: []*gltf.PolyformModel{{Name: label, Mesh: &mesh}}, rendered: label + " (mesh, default gray material)"}, nil
	case nodes.Output[sample.Vec3ToFloat]:
		mesh, march, err := marchForPreview(out.Value(), nil, in)
		if err != nil {
			return renderSubject{}, fmt.Errorf("%s: %w", label, err)
		}
		return renderSubject{models: []*gltf.PolyformModel{{Name: label, Mesh: &mesh}}, rendered: label + " (SDF field marched for this render, default gray material)", march: march}, nil
	case nodes.Output[sdf.ColoredField]:
		cf := out.Value()
		if cf.Distance == nil {
			return renderSubject{}, fmt.Errorf("%s is an empty colored field", label)
		}
		mesh, march, err := marchForPreview(cf.Distance, cf.Color, in)
		if err != nil {
			return renderSubject{}, fmt.Errorf("%s: %w", label, err)
		}
		return renderSubject{models: []*gltf.PolyformModel{{Name: label, Mesh: &mesh}}, rendered: label + " (colored SDF field marched for this render, its own colors)", march: march}, nil
	}
	typeName := "an unknown type"
	if typed, ok := outputs[name].(nodes.Typed); ok {
		typeName = typed.Type()
	}
	return renderSubject{}, fmt.Errorf("%s is %s; render_preview can show a glTF Model, a Mesh or an SDF field (sample.Vec3ToFloat or sdf.ColoredField)", label, typeName)
}

func renderableOutputs(outputs map[string]nodes.OutputPort) []string {
	var found []string
	for _, name := range sortedOutputKeys(outputs) {
		switch outputs[name].(type) {
		case nodes.Output[*gltf.PolyformModel], nodes.Output[[]*gltf.PolyformModel], nodes.Output[modeling.Mesh], nodes.Output[sample.Vec3ToFloat], nodes.Output[sdf.ColoredField]:
			found = append(found, name)
		}
	}
	return found
}

const (
	previewMarchCellsAcross = 80
	fieldBoundsCells        = 24
)

func marchForPreview(field sample.Vec3ToFloat, paint sdf.ColorField, in RenderPreviewInput) (modeling.Mesh, *RenderMarch, error) {
	var domain geometry.AABB
	if in.Domain != nil {
		domain = geometry.NewAABB(
			vector3.New(in.Domain.Center.X, in.Domain.Center.Y, in.Domain.Center.Z),
			vector3.New(in.Domain.Size.X, in.Domain.Size.Y, in.Domain.Size.Z),
		)
	} else {
		var err error
		if domain, err = estimateFieldBounds(field); err != nil {
			return modeling.Mesh{}, nil, err
		}
	}
	size := domain.Size()
	longest := math.Max(size.X(), math.Max(size.Y(), size.Z()))
	if longest <= 0 {
		return modeling.Mesh{}, nil, fmt.Errorf("the march domain has no size")
	}
	resolution := in.MarchResolution
	if resolution <= 0 {
		resolution = previewMarchCellsAcross / longest
	}
	cube := 1 / resolution

	mesh := marching.March(field, domain, cube, 0)
	march := &RenderMarch{
		Domain:     FieldProbeDomain{Center: toVector3Input(domain.Center()), Size: toVector3Input(size)},
		Resolution: resolution,
		CubeSize:   cube,
	}
	if !mesh.HasFloat3Attribute(modeling.PositionAttribute) || mesh.PrimitiveCount() == 0 {
		return modeling.Mesh{}, march, fmt.Errorf("marching the field over a %.3g x %.3g x %.3g domain at %.3g cells per unit found no surface: the field never crosses zero inside it, or the domain misses the shape (pass domain)", size.X(), size.Y(), size.Z(), resolution)
	}
	mesh = meshops.SmoothNormals(mesh)
	if paint != nil {
		positions := mesh.Float3Attribute(modeling.PositionAttribute)
		colors := make([]vector3.Float64, positions.Len())
		for i := range colors {
			c := paint(positions.At(i))
			colors[i] = vector3.New(c.R, c.G, c.B)
		}
		mesh = mesh.SetFloat3Attribute(modeling.ColorAttribute, colors)
	}
	march.Triangles = mesh.PrimitiveCount()
	return mesh, march, nil
}

func toVector3Input(v vector3.Float64) Vector3Input {
	return Vector3Input{X: v.X(), Y: v.Y(), Z: v.Z()}
}

// estimateFieldBounds boxes where the field is at or near zero. A coarse
// grid over a cube at the origin doubles until the near-surface samples
// stop touching its faces, then the box is re-sampled to tighten it.
func estimateFieldBounds(field sample.Vec3ToFloat) (geometry.AABB, error) {
	var box geometry.AABB
	found := false
	for half := 1.; half <= 128; half *= 2 {
		probe := geometry.NewAABB(vector3.Zero[float64](), vector3.New(2*half, 2*half, 2*half))
		candidate, ok, touching := nearSurfaceBox(field, probe)
		if ok && !touching {
			box, found = candidate, true
			break
		}
	}
	if !found {
		return geometry.AABB{}, fmt.Errorf("couldn't bound the field: within 128 units of the origin it either has no surface or never stops (an unbounded plane, or negative everywhere); pass domain to say where to march")
	}
	for range 3 {
		tighter, ok, _ := nearSurfaceBox(field, box)
		if !ok {
			break
		}
		box = tighter
	}
	return geometry.NewAABB(box.Center(), box.Size().Scale(1.05)), nil
}

// nearSurfaceBox samples a grid over probe and boxes the samples within half
// a cell diagonal of the surface, one cell wider on every side. touching
// says such a sample lay on the probe's outer layer, so the surface may go
// on past it.
func nearSurfaceBox(field sample.Vec3ToFloat, probe geometry.AABB) (geometry.AABB, bool, bool) {
	step := probe.Size().Scale(1. / fieldBoundsCells)
	near := step.Length() / 2
	origin := probe.Min()
	var points []vector3.Float64
	touching := false
	for i := 0; i <= fieldBoundsCells; i++ {
		for j := 0; j <= fieldBoundsCells; j++ {
			for k := 0; k <= fieldBoundsCells; k++ {
				p := origin.Add(vector3.New(step.X()*float64(i), step.Y()*float64(j), step.Z()*float64(k)))
				v := field(p)
				if math.IsNaN(v) || v > near {
					continue
				}
				points = append(points, p)
				if i == 0 || j == 0 || k == 0 || i == fieldBoundsCells || j == fieldBoundsCells || k == fieldBoundsCells {
					touching = true
				}
			}
		}
	}
	if len(points) == 0 {
		return geometry.AABB{}, false, false
	}
	tight := geometry.NewAABBFromPoints(points...)
	return geometry.NewAABB(tight.Center(), tight.Size().Add(step.Scale(2))), true, touching
}

// flattenModel walks a model tree composing parent*local TRS and bakes the
// world transform into a copy of each mesh, one part per GPU instance.
// Instances repeat the model's own mesh before its TRS and never its
// children: EXT_mesh_gpu_instancing draws it so.
func flattenModel(model *gltf.PolyformModel, parent mat.Matrix4x4, lineage []*gltf.PolyformModel, out *[]renderablePart) {
	lineage = append(lineage[:len(lineage):len(lineage)], model)

	world := parent
	if model.TRS != nil {
		world = parent.Multiply(model.TRS.Matrix())
	}

	// Transforming a mesh with no Position attribute panics deep in meshops
	// without naming the model, so an empty mesh is skipped here.
	if model.Mesh != nil && model.Mesh.HasFloat3Attribute(modeling.PositionAttribute) {
		placements := []mat.Matrix4x4{world}
		if len(model.GpuInstances) > 0 {
			placements = placements[:0]
			for _, instance := range model.GpuInstances {
				placements = append(placements, world.Multiply(instance.Matrix()))
			}
		}
		color, explicit := materialColor(model.Material)
		for _, placement := range placements {
			*out = append(*out, renderablePart{
				Mesh:          bakeTransform(*model.Mesh, placement),
				Color:         color,
				ExplicitColor: explicit,
				Texture:       materialTexture(model.Material),
				Finish:        materialFinish(model.Material),
				Lineage:       lineage,
			})
		}
	}

	for _, child := range model.Children {
		flattenModel(child, world, lineage, out)
	}
}

// buildScene flattens the models into draw groups, dropping the parts that
// exclude names and, when only is set, every part it does not.
func buildScene(models []*gltf.PolyformModel, exclude, only modelRefs, nodeID string, excluded int) (renderScene, error) {
	var parts []renderablePart
	for _, model := range models {
		flattenModel(model, mat.Identity(), nil, &parts)
	}

	var scene renderScene
	warned := map[string]bool{}
	for _, part := range parts {
		if exclude.matches(part.Lineage) || (!only.empty() && !only.matches(part.Lineage)) {
			scene.parts = append(scene.parts, scenePart{lineage: part.Lineage})
			continue
		}
		triangles := partTriangles(part)
		if len(triangles) == 0 {
			continue
		}
		if bad, total := unusableNormals(part.Mesh); bad > 0 {
			path := lineagePath(part.Lineage)
			if !warned[path] {
				warned[path] = true
				scene.warnings = append(scene.warnings, fmt.Sprintf("%s: %d of %d vertex normals are zero or not finite", path, bad, total))
			}
		}
		scene.triangles = append(scene.triangles, triangles...)
		scene.parts = append(scene.parts, scenePart{triangles: triangles, lineage: part.Lineage})
		group := drawGroup{Mesh: fauxgl.NewTriangleMesh(triangles), Finish: part.Finish}
		if part.Texture != nil && part.Mesh.HasFloat2Attribute(modeling.TexCoordAttribute) {
			group.Texture = part.Texture
		}
		scene.groups = append(scene.groups, group)
	}
	if len(scene.triangles) == 0 {
		if !only.empty() || excluded > 0 {
			return scene, fmt.Errorf("nothing is left to render feeding node %q once exclude (%d) and only are applied", nodeID, excluded)
		}
		return scene, fmt.Errorf("no renderable meshes found feeding node %q. Every model reaching it is empty - if any come from a March node, that usually means its Resolution is too low for its Domain (voxel size is 1/Resolution, and must be smaller than the domain) or the field never crosses the surface inside the domain", nodeID)
	}
	return scene, nil
}

// partTriangles builds a part's fauxgl triangles. glTF multiplies COLOR_0 by
// baseColorFactor rather than letting either win, and the texture by both,
// so a vertex-painted or textured part is combined the way a viewer draws
// it; a part with neither takes the material's flat color.
func partTriangles(part renderablePart) []*fauxgl.Triangle {
	vertexColored := part.Mesh.HasFloat3Attribute(modeling.ColorAttribute)
	textured := part.Texture != nil && part.Mesh.HasFloat2Attribute(modeling.TexCoordAttribute)
	smooth := part.Mesh.HasFloat3Attribute(modeling.NormalAttribute)

	tint := fauxgl.Color{R: 1, G: 1, B: 1, A: 1}
	if part.ExplicitColor {
		tint = part.Color
	}

	triangles := make([]*fauxgl.Triangle, 0, part.Mesh.PrimitiveCount())
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
		if smooth {
			t.V1.Normal = vertexNormal(tri.P1Vec3Attr(modeling.NormalAttribute))
			t.V2.Normal = vertexNormal(tri.P2Vec3Attr(modeling.NormalAttribute))
			t.V3.Normal = vertexNormal(tri.P3Vec3Attr(modeling.NormalAttribute))
		}
		switch {
		case textured:
			uv1 := tri.P1Vec2Attr(modeling.TexCoordAttribute)
			uv2 := tri.P2Vec2Attr(modeling.TexCoordAttribute)
			uv3 := tri.P3Vec2Attr(modeling.TexCoordAttribute)
			t.V1.Texture = fauxgl.V(uv1.X(), uv1.Y(), 0)
			t.V2.Texture = fauxgl.V(uv2.X(), uv2.Y(), 0)
			t.V3.Texture = fauxgl.V(uv3.X(), uv3.Y(), 0)

			c1, c2, c3 := tint, tint, tint
			if vertexColored {
				c1 = tintedVertexColor(tri.P1Vec3Attr(modeling.ColorAttribute), tint)
				c2 = tintedVertexColor(tri.P2Vec3Attr(modeling.ColorAttribute), tint)
				c3 = tintedVertexColor(tri.P3Vec3Attr(modeling.ColorAttribute), tint)
			}
			t.V1.Color, t.V2.Color, t.V3.Color = c1, c2, c3
		case vertexColored:
			t.V1.Color = tintedVertexColor(tri.P1Vec3Attr(modeling.ColorAttribute), tint)
			t.V2.Color = tintedVertexColor(tri.P2Vec3Attr(modeling.ColorAttribute), tint)
			t.V3.Color = tintedVertexColor(tri.P3Vec3Attr(modeling.ColorAttribute), tint)
		default:
			t.V1.Color = part.Color
			t.V2.Color = part.Color
			t.V3.Color = part.Color
		}
		triangles = append(triangles, t)
	}
	return triangles
}
