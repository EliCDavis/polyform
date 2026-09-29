package extrude

import (
	"errors"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Outline struct {
	Shape  []vector2.Float64
	Path   []vector3.Float64
	Closed bool

	// Outline corners turning at least this many degrees keep a hard edge.
	// Nil smooths anything under defaultSmoothingAngle.
	SmoothingAngle *float64

	// Lays the sweep out as a strip: across is the outline's perimeter,
	// along is the path. Nil writes no texture coordinates at all.
	UVs *primitives.StripUVs
}

const defaultSmoothingAngle = 45.

func (o Outline) smoothing() float64 {
	if o.SmoothingAngle == nil {
		return defaultSmoothingAngle * math.Pi / 180
	}
	return *o.SmoothingAngle * math.Pi / 180
}

type outlineSlot struct {
	point  int
	normal vector2.Float64
}

// A ring only needs two vertices at an outline corner where the outline
// actually turns one. Splitting every vertex leaves a spline sampled outline
// carrying twice the vertices it needs and a seam of split normals at every
// sample, which reads as faceting rather than a smooth sweep.
//
// splitSeam gives the loop's closing point two slots even where the outline
// runs smoothly through it, so one can carry perimeter 0 and the other 1.
// Both keep the averaged normal, so the split shows in the UVs and not in
// the shading.
func outlineSlots(shape []vector2.Float64, smoothing float64, splitSeam bool) (slots []outlineSlot, edges []vector2.Float64, start, end []int) {
	count := len(shape)
	edges = make([]vector2.Float64, count)
	for j := range shape {
		along := shape[(j+1)%count].Sub(shape[j])
		if along.Length() == 0 {
			continue
		}
		edges[j] = vector2.New(along.Y(), -along.X()).Normalized()
	}

	start = make([]int, count)
	end = make([]int, count)

	for j := range shape {
		previous := (j + count - 1) % count
		incoming, outgoing := edges[previous], edges[j]

		switch {
		case incoming.Length() == 0:
			incoming = outgoing
		case outgoing.Length() == 0:
			outgoing = incoming
		}

		// Nudged so a shape whose corners all sit exactly on the threshold,
		// a regular octagon against the 45 degree default, resolves the same
		// way at every one of them instead of splitting half of them.
		if incoming.Length() > 0 && incoming.Angle(outgoing) < smoothing-1e-9 {
			averaged := incoming.Add(outgoing).Normalized()
			if splitSeam && j == 0 {
				slots = append(slots, outlineSlot{j, averaged})
				end[previous] = len(slots) - 1
				slots = append(slots, outlineSlot{j, averaged})
				start[j] = len(slots) - 1
				continue
			}
			slots = append(slots, outlineSlot{j, averaged})
			start[j] = len(slots) - 1
			end[previous] = len(slots) - 1
			continue
		}

		slots = append(slots, outlineSlot{j, incoming})
		end[previous] = len(slots) - 1
		slots = append(slots, outlineSlot{j, outgoing})
		start[j] = len(slots) - 1
	}

	return
}

// The shell faces outward only for a counter clockwise outline, and callers
// have no reason to know that.
func counterClockwise(shape []vector2.Float64) []vector2.Float64 {
	if geometry.Shape(shape).SignedArea() >= 0 {
		return shape
	}

	flipped := make([]vector2.Float64, len(shape))
	for i, p := range shape {
		flipped[len(shape)-1-i] = p
	}
	return flipped
}

// A point repeating its predecessor, or the first point repeated at the end
// to close the loop, adds an edge of no length: a zero-area quad in the
// shell, and a NaN normal where two such edges meet.
func withoutRepeats(shape []vector2.Float64) []vector2.Float64 {
	kept := make([]vector2.Float64, 0, len(shape))
	for _, p := range shape {
		if len(kept) > 0 && p == kept[len(kept)-1] {
			continue
		}
		kept = append(kept, p)
	}
	for len(kept) > 1 && kept[len(kept)-1] == kept[0] {
		kept = kept[:len(kept)-1]
	}
	return kept
}

func finite(v vector3.Float64) bool {
	return !math.IsNaN(v.X()) && !math.IsNaN(v.Y()) && !math.IsNaN(v.Z()) &&
		!math.IsInf(v.X(), 0) && !math.IsInf(v.Y(), 0) && !math.IsInf(v.Z(), 0)
}

// ringBasis matches the frame ProjectFace builds, so shell and cap vertices
// at the same path point land on top of each other instead of leaving a seam.
type ringBasis struct {
	origin, dir, per, cross vector3.Float64
}

func basisAt(path []vector3.Float64, frames []ringFrame, i int) ringBasis {
	return ringBasis{
		origin: path[i],
		dir:    frames[i].dir,
		per:    frames[i].per,
		cross:  frames[i].dir.Cross(frames[i].per),
	}
}

func (b ringBasis) place(p vector2.Float64) vector3.Float64 {
	return b.origin.Add(b.cross.Scale(p.X())).Add(b.per.Scale(p.Y()))
}

func (b ringBasis) direction(p vector2.Float64) vector3.Float64 {
	return b.cross.Scale(p.X()).Add(b.per.Scale(p.Y()))
}

// ConstrainedDelaunay keeps input points at their original indices, so the
// first len(shape) cap vertices line up with the shell ring.
func tessellate(shape []vector2.Float64) ([]vector2.Float64, []int, error) {
	mesh, err := triangulation.ConstrainedDelaunay(
		shape, []triangulation.Constraint{triangulation.NewConstraint(shape)})
	if err != nil {
		return nil, nil, err
	}

	pos := mesh.Float3Attribute(modeling.PositionAttribute)
	points := make([]vector2.Float64, pos.Len())
	for i := 0; i < pos.Len(); i++ {
		v := pos.At(i)
		points[i] = vector2.New(v.X(), v.Z())
	}

	idx := mesh.Indices()
	tris := make([]int, idx.Len())
	for i := 0; i < idx.Len(); i++ {
		tris[i] = idx.At(i)
	}

	return points, tris, nil
}

func perimeterFractions(shape []vector2.Float64, slots []outlineSlot, seam int) []float64 {
	count := len(shape)
	walked := make([]float64, count+1)
	for j := 0; j < count; j++ {
		walked[j+1] = walked[j] + shape[(j+1)%count].Sub(shape[j]).Length()
	}

	fractions := make([]float64, len(slots))
	if perimeter := walked[count]; perimeter > 0 {
		for i, slot := range slots {
			fractions[i] = walked[slot.point] / perimeter
		}
		if seam >= 0 && seam < len(fractions) {
			fractions[seam] = 1
		}
	}
	return fractions
}

func pathFractions(path []vector3.Float64, rings int) []float64 {
	walked := make([]float64, rings)
	for i := 1; i < rings; i++ {
		walked[i] = walked[i-1] + path[i%len(path)].Sub(path[(i-1)%len(path)]).Length()
	}

	total := walked[rings-1]
	if total == 0 {
		return walked
	}
	for i := range walked {
		walked[i] /= total
	}
	return walked
}

func (o Outline) shell(shape []vector2.Float64, frames []ringFrame) modeling.Mesh {
	slots, edges, start, end := outlineSlots(shape, o.smoothing(), o.UVs != nil)
	perRing := len(slots)

	rings := len(o.Path)
	if o.UVs != nil && o.Closed {
		rings++
	}

	var across, along []float64
	var uvs []vector2.Float64
	if o.UVs != nil {
		across = perimeterFractions(shape, slots, end[len(shape)-1])
		along = pathFractions(o.Path, rings)
		uvs = make([]vector2.Float64, 0, rings*perRing)
	}

	vertices := make([]vector3.Float64, 0, rings*perRing)
	normals := make([]vector3.Float64, 0, rings*perRing)

	for i := 0; i < rings; i++ {
		basis := basisAt(o.Path, frames, i%len(o.Path))
		for s, slot := range slots {
			vertices = append(vertices, basis.place(shape[slot.point]))
			normals = append(normals, basis.direction(slot.normal).Normalized())
			if o.UVs != nil {
				uvs = append(uvs, o.UVs.AtXY(vector2.New(across[s], along[i])))
			}
		}
	}

	segments := len(o.Path) - 1
	if o.Closed {
		segments = len(o.Path)
	}

	tris := make([]int, 0, segments*len(shape)*6)
	for i := 0; i < segments; i++ {
		lower := i * perRing
		upper := ((i + 1) % rings) * perRing
		basis := basisAt(o.Path, frames, i)

		for j := range shape {
			a, b := lower+start[j], lower+end[j]
			c, d := upper+start[j], upper+end[j]

			// The slot normal may be averaged across a smooth corner, so the
			// winding is judged against the edge's own outward direction.
			outward := basis.direction(edges[j])
			face := vertices[c].Sub(vertices[a]).Cross(vertices[b].Sub(vertices[a]))
			if face.Dot(outward) < 0 {
				tris = append(tris, a, b, d, a, d, c)
			} else {
				tris = append(tris, a, c, d, a, d, b)
			}
		}
	}

	mesh := modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		})
	if o.UVs == nil {
		return mesh
	}
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

func (o Outline) cap(points []vector2.Float64, tris []int, frames []ringFrame, at int, facing float64) modeling.Mesh {
	basis := basisAt(o.Path, frames, at)
	normal := basis.dir.Scale(facing)

	vertices := make([]vector3.Float64, len(points))
	normals := make([]vector3.Float64, len(points))
	for i, p := range points {
		vertices[i] = basis.place(p)
		normals[i] = normal
	}

	wound := make([]int, 0, len(tris))
	for i := 0; i < len(tris); i += 3 {
		a, b, c := tris[i], tris[i+1], tris[i+2]
		if vertices[b].Sub(vertices[a]).Cross(vertices[c].Sub(vertices[a])).Dot(normal) < 0 {
			b, c = c, b
		}
		wound = append(wound, a, b, c)
	}

	mesh := modeling.NewTriangleMesh(wound).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		})
	if o.UVs == nil {
		return mesh
	}

	// The strip has no room for a cap, so each one takes the whole of it,
	// the profile's bounding box stretched over the strip's quad.
	min, max := points[0], points[0]
	for _, p := range points {
		min = vector2.New(math.Min(min.X(), p.X()), math.Min(min.Y(), p.Y()))
		max = vector2.New(math.Max(max.X(), p.X()), math.Max(max.Y(), p.Y()))
	}
	size := max.Sub(min)

	uvs := make([]vector2.Float64, len(points))
	for i, p := range points {
		unit := p.Sub(min)
		if size.X() > 0 {
			unit = vector2.New(unit.X()/size.X(), unit.Y())
		}
		if size.Y() > 0 {
			unit = vector2.New(unit.X(), unit.Y()/size.Y())
		}
		uvs[i] = o.UVs.AtXY(unit)
	}
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

func (o Outline) Extrude() (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)

	shape := withoutRepeats(o.Shape)
	if len(shape) < 3 {
		return empty, fmt.Errorf("outline needs at least 3 distinct points, got %d", len(shape))
	}
	if len(o.Path) < 2 {
		return empty, fmt.Errorf("path needs at least 2 points, got %d", len(o.Path))
	}
	if geometry.Shape(shape).SignedArea() == 0 {
		return empty, errors.New("outline encloses no area")
	}

	// A zero length segment does not always poison the frame, so it has to
	// be caught outright. A closed path repeating its first point at the end
	// is the common way in, and it silently lands two rings on top of oneother.
	segments := len(o.Path) - 1
	if o.Closed {
		segments = len(o.Path)
	}
	for i := 0; i < segments; i++ {
		next := (i + 1) % len(o.Path)
		if o.Path[next].Sub(o.Path[i]).Length() == 0 {
			return empty, fmt.Errorf("path points %d and %d are the same point", i, next)
		}
	}

	frames := pathFrames(o.Path, o.Closed)
	for i, f := range frames {
		if !finite(f.dir) || !finite(f.per) {
			return empty, fmt.Errorf("path has no usable frame at point %d; it doubles back on itself", i)
		}
	}

	shape = counterClockwise(shape)
	shell := o.shell(shape, frames)
	if o.Closed {
		return shell, nil
	}

	points, tris, err := tessellate(shape)
	if err != nil {
		return empty, err
	}

	return shell.
		Append(o.cap(points, tris, frames, 0, -1)).
		Append(o.cap(points, tris, frames, len(o.Path)-1, 1)), nil
}

type OutlineNode struct {
	Outline        nodes.Output[[]vector2.Float64]   `description:"Closed 2D outline to sweep. Winding does not matter. Its frame is set by the path direction: swept along +Y, x lies on +X and y on +Z; along -Z (toward the viewer), x on +X and y on +Y; along +Z, x on -X (mirrored) and y on +Y; along +X, x on -Y and y on +Z."`
	Path           nodes.Output[[]vector3.Float64]   `description:"Points the outline is swept through. For a flat plate facing +Z, run the path from +Z toward -Z so the outline's x keeps world +X."`
	Closed         nodes.Output[bool]                `description:"Joins the last path point back to the first and leaves the ends uncapped."`
	SmoothingAngle nodes.Output[float64]             `description:"Outline corners turning at least this many degrees keep a hard edge. Defaults to 45."`
	UVs            nodes.Output[primitives.StripUVs] `description:"Lays the sweep out as a strip: the outline's perimeter runs across it, the path along it. Each cap takes the whole strip separately. Unconnected, the mesh carries no texture coordinates at all."`
}

func (node OutlineNode) Description() string {
	return "Sweeps a 2D outline along a path, capping each end with a triangulation of the outline."
}

func (node OutlineNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))

	if node.Outline == nil || node.Path == nil {
		return
	}

	smoothing := nodes.TryGetOutputValue(out, node.SmoothingAngle, defaultSmoothingAngle)
	var uvs *primitives.StripUVs
	if node.UVs != nil {
		strip := nodes.GetOutputValue(out, node.UVs)
		uvs = &strip
	}
	mesh, err := Outline{
		Shape:          nodes.GetOutputValue(out, node.Outline),
		Path:           nodes.GetOutputValue(out, node.Path),
		Closed:         nodes.TryGetOutputValue(out, node.Closed, false),
		SmoothingAngle: &smoothing,
		UVs:            uvs,
	}.Extrude()
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}
