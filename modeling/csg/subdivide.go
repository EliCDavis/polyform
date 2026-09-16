package csg

import (
	"errors"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Section 3 holds one array of vertices with polygons pointing into it, so
// which faces meet is known without asking where their corners are. Carrying
// only corners means anything needing that has to recover it by position.
type face struct {
	verts  [3]vector3.Float64
	normal vector3.Float64
}

// A triangle enclosing no area contributes no surface, and its normal is
// either undefined or, for one thin enough to underflow, junk. Section 7
// steers its ray by that normal, so these are dropped rather than carried.
func newFace(a, b, c vector3.Float64) (face, bool) {
	normal := b.Sub(a).Cross(c.Sub(a))
	if normal.Length() == 0 {
		return face{}, false
	}

	unit := normal.Normalized()
	for _, component := range []float64{unit.X(), unit.Y(), unit.Z()} {
		if math.IsNaN(component) || math.IsInf(component, 0) {
			return face{}, false
		}
	}

	return face{
		verts:  [3]vector3.Float64{a, b, c},
		normal: unit,
	}, true
}

func (f face) reversed() face {
	return face{
		verts:  [3]vector3.Float64{f.verts[0], f.verts[2], f.verts[1]},
		normal: f.normal.Scale(-1),
	}
}

// Section 7 calls the average of a polygon's vertices its barycenter.
func (f face) barycenter() vector3.Float64 {
	return f.verts[0].Add(f.verts[1]).Add(f.verts[2]).Scale(1. / 3.)
}

func (f face) bounds() geometry.AABB {
	return geometry.NewAABBFromPoints(f.verts[0], f.verts[1], f.verts[2])
}

// How far the face reaches from its own barycenter. Half the diagonal of its
// bounding box is not this: a right triangle on the unit axes puts a corner
// 0.745 out while that reads 0.707, and a range query cut to the shorter one
// quietly misses points sitting on the face's own edges.
func (f face) reach() float64 {
	center := f.barycenter()
	return math.Max(
		f.verts[0].Sub(center).Length(),
		math.Max(
			f.verts[1].Sub(center).Length(),
			f.verts[2].Sub(center).Length(),
		),
	)
}

// Which side of other's plane each of this face's corners falls on.
//
// Not a distance: a determinant, exact in sign, scaled by twice other's area.
// Callers only read its sign and the ratio of two of them, both of which that
// scaling leaves alone.
func (f face) distancesTo(other face) [3]float64 {
	var out [3]float64
	for i, v := range f.verts {
		out[i] = predicate.Orient3D(other.verts[0], other.verts[1], other.verts[2], v)
	}
	return out
}

type segment [2]vector3.Float64

type faceElement struct {
	bounds geometry.AABB
}

func (e faceElement) BoundingBox() geometry.AABB { return e.bounds }

func (e faceElement) ClosestPoint(p vector3.Float64) vector3.Float64 {
	return e.bounds.ClosestPoint(p)
}

func facesOf(m modeling.Mesh) ([]face, error) {
	if m.Topology() != modeling.TriangleTopology {
		return nil, fmt.Errorf("need a triangle mesh, got %v", m.Topology())
	}
	if !m.HasFloat3Attribute(modeling.PositionAttribute) {
		return nil, errors.New("mesh carries no position attribute")
	}

	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	faces := make([]face, 0, indices.Len()/3)
	for i := 0; i+2 < indices.Len(); i += 3 {
		f, ok := newFace(
			positions.At(indices.At(i)),
			positions.At(indices.At(i+1)),
			positions.At(indices.At(i+2)),
		)
		if ok {
			faces = append(faces, f)
		}
	}
	return faces, nil
}

func meshOf(faces []face) modeling.Mesh {
	tris := make([]int, 0, len(faces)*3)
	verts := make([]vector3.Float64, 0, len(faces)*3)
	normals := make([]vector3.Float64, 0, len(faces)*3)

	for _, f := range faces {
		tris = append(tris, len(verts), len(verts)+1, len(verts)+2)
		verts = append(verts, f.verts[0], f.verts[1], f.verts[2])
		normals = append(normals, f.normal, f.normal, f.normal)
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: verts,
			modeling.NormalAttribute:   normals,
		})
}

// Every comparison against zero here is really a comparison against the size
// of what is being cut, so a millimetre model and a kilometre one behave the
// same.
func toleranceFor(a, b []face) float64 {
	low := vector3.Fill(math.Inf(1))
	high := vector3.Fill(math.Inf(-1))

	// Walked in place. Copying both face slices into one to range over cost
	// more memory than everything else in the operation put together.
	stretch := func(faces []face) {
		for _, f := range faces {
			for _, v := range f.verts {
				low = vector3.New(
					math.Min(low.X(), v.X()),
					math.Min(low.Y(), v.Y()),
					math.Min(low.Z(), v.Z()))
				high = vector3.New(
					math.Max(high.X(), v.X()),
					math.Max(high.Y(), v.Y()),
					math.Max(high.Z(), v.Z()))
			}
		}
	}
	stretch(a)
	stretch(b)

	span := high.Sub(low).Length()
	if span == 0 || math.IsInf(span, 0) || math.IsNaN(span) {
		return 1e-9
	}
	return span * 1e-9
}

// Section 5, "Do Two Polygons Intersect?": the segment shared by two faces.
//
// Reports false for coplanar pairs. Section 4's outer loop leaves those
// alone, expecting the faces around them to do the cutting.
func sharedSegment(a, b face, tolerance float64) (segment, bool) {
	toB := a.distancesTo(b)
	toA := b.distancesTo(a)

	if entirelyOneSide(toB) || entirelyOneSide(toA) {
		return segment{}, false
	}
	if coplanar(toB) {
		return segment{}, false
	}

	onB, ok := crossesPlane(a, toB, tolerance)
	if !ok {
		return segment{}, false
	}
	onA, ok := crossesPlane(b, toA, tolerance)
	if !ok {
		return segment{}, false
	}

	// Both segments lie on the line where the two planes meet, so they can be
	// compared as intervals along it.
	direction := a.normal.Cross(b.normal)
	if direction.Length() == 0 {
		return segment{}, false
	}
	direction = direction.Normalized()

	base := onB[0]
	at := func(p vector3.Float64) float64 { return p.Sub(base).Dot(direction) }

	// The overlap always ends on one of the four crossing points already in
	// hand, so the answer is which one, not where. Rebuilding it from a
	// parameter would round it through a normalized direction and a scale,
	// and the same point reached from the other face would land elsewhere -
	// which is the drift the snapping tolerances downstream exist to absorb.
	ends := func(p [2]vector3.Float64) (low, high vector3.Float64, lowAt, highAt float64) {
		first, second := at(p[0]), at(p[1])
		if first <= second {
			return p[0], p[1], first, second
		}
		return p[1], p[0], second, first
	}

	aLow, aHigh, aLowAt, aHighAt := ends(onB)
	bLow, bHigh, bLowAt, bHighAt := ends(onA)

	low, lowAt := aLow, aLowAt
	if bLowAt > lowAt {
		low, lowAt = bLow, bLowAt
	}

	high, highAt := aHigh, aHighAt
	if bHighAt < highAt {
		high, highAt = bHigh, bHighAt
	}

	if highAt-lowAt <= tolerance {
		return segment{}, false
	}

	return segment{low, high}, true
}

// Compared against zero rather than a tolerance: sideOfPlane is exact, so
// there is no band of uncertainty left to allow for.
func entirelyOneSide(distances [3]float64) bool {
	positive, negative := 0, 0
	for _, d := range distances {
		if d > 0 {
			positive++
		}
		if d < 0 {
			negative++
		}
	}
	return positive == 3 || negative == 3
}

// Whether a lies on b's plane closely enough that classification will call it
// a boundary face. coplanar is exact and answers a different question: it
// decides whether to cut, where being a rounding error off the plane really
// does mean the two faces cross.
func sharesPlane(a, b face, tolerance float64) bool {
	for _, v := range a.verts {
		if math.Abs(b.planeDistance(v)) > tolerance {
			return false
		}
	}
	return true
}

func coplanar(distances [3]float64) bool {
	for _, d := range distances {
		if d != 0 {
			return false
		}
	}
	return true
}

// Where a face meets a plane: two points, or nothing when it only grazes a
// single corner.
func crossesPlane(f face, distances [3]float64, tolerance float64) ([2]vector3.Float64, bool) {
	found := make([]vector3.Float64, 0, 2)

	add := func(p vector3.Float64) {
		for _, existing := range found {
			if existing.Sub(p).Length() <= tolerance {
				return
			}
		}
		found = append(found, p)
	}

	for i := 0; i < 3; i++ {
		j := (i + 1) % 3
		from, to := distances[i], distances[j]

		if from == 0 {
			add(f.verts[i])
			continue
		}
		if to == 0 || (from > 0) == (to > 0) {
			continue
		}

		// Both are determinants carrying the same area factor, so it cancels
		// and the ratio is the true fraction along the edge.

		along := from / (from - to)
		add(f.verts[i].Add(f.verts[j].Sub(f.verts[i]).Scale(along)))
	}

	if len(found) != 2 {
		return [2]vector3.Float64{}, false
	}
	return [2]vector3.Float64{found[0], found[1]}, true
}

// Section 4, "Intersecting the Objects": where the faces of against cross
// each face of target.
//
// The endpoints come back alongside the cuts because both solids have to be
// split at all of them, not just at their own. The curve where the two meet
// belongs to both, and a point only one of them splits at leaves the other
// with an edge running straight past it.
// touching reports the target faces that share a plane with a face of the
// other solid. Section 4 leaves such pairs unsplit, so no cut ever separates
// them from their neighbours even though they classify differently.
func cutsAgainst(target, against []face, tolerance float64) (cuts [][]segment, corners []vector3.Float64, touching []bool) {
	elements := make([]trees.Element, len(against))
	for i, f := range against {
		elements[i] = faceElement{bounds: f.bounds()}
	}
	tree := trees.NewOctree(elements)

	cuts = make([][]segment, len(target))
	corners = make([]vector3.Float64, 0)
	touching = make([]bool, len(target))

	for i, f := range target {
		reach := f.reach()
		for _, j := range tree.ElementsWithinRange(f.barycenter(), reach+tolerance) {
			cut, ok := sharedSegment(f, against[j], tolerance)
			if !ok {
				if sharesPlane(f, against[j], tolerance) {
					touching[i] = true
				}
				continue
			}
			cuts[i] = append(cuts[i], cut)
			corners = append(corners, cut[0], cut[1])
		}
	}

	return cuts, corners, touching
}

// Splits every face along its own cuts, and at any corner from either solid
// that happens to land on it.
func splitAll(
	target []face,
	ids [][3]int,
	cuts [][]segment,
	corners []vector3.Float64,
	touching []bool,
	tolerance float64,
	weld *welder,
) ([]face, [][3]int, []bool, error) {
	if len(corners) == 0 {
		return target, ids, touching, nil
	}

	elements := make([]trees.Element, len(corners))
	for i, c := range corners {
		elements[i] = faceElement{bounds: geometry.NewAABBFromPoints(c)}
	}
	tree := trees.NewOctree(elements)

	out := make([]face, 0, len(target))
	outIDs := make([][3]int, 0, len(target))
	outTouching := make([]bool, 0, len(target))

	for i, f := range target {
		reach := f.reach()
		landed := make([]vector3.Float64, 0)
		for _, j := range tree.ElementsWithinRange(f.barycenter(), reach+tolerance) {
			if onFace(f, corners[j], tolerance) {
				landed = append(landed, corners[j])
			}
		}

		// An uncut face keeps the ids it came in with, which is most of them.
		if len(cuts[i]) == 0 && len(landed) == 0 {
			out = append(out, f)
			outIDs = append(outIDs, ids[i])
			outTouching = append(outTouching, touching[i])
			continue
		}

		pieces, pieceIDs, err := splitFace(f, ids[i], cuts[i], landed, tolerance, weld)
		if err != nil {
			return nil, nil, nil, err
		}
		out = append(out, pieces...)
		outIDs = append(outIDs, pieceIDs...)
		for range pieces {
			outTouching = append(outTouching, touching[i])
		}
	}

	return out, outIDs, outTouching, nil
}

func onFace(f face, p vector3.Float64, tolerance float64) bool {
	if math.Abs(p.Sub(f.verts[0]).Dot(f.normal)) > tolerance {
		return false
	}
	for i := 0; i < 3; i++ {
		j := (i + 1) % 3
		edge := f.verts[j].Sub(f.verts[i])
		if edge.Cross(p.Sub(f.verts[i])).Dot(f.normal) < -tolerance*edge.Length() {
			return false
		}
	}
	return true
}

// Section 6, "Subdividing Non-coplanar Polygons".
//
// The paper walks a case analysis over where the segment meets the polygon so
// the pieces stay convex. Triangles fed to a constrained triangulator come
// out convex anyway, so the cuts are handed over as edges it has to keep.
func splitFace(
	f face,
	ids [3]int,
	cuts []segment,
	landed []vector3.Float64,
	tolerance float64,
	weld *welder,
) ([]face, [][3]int, error) {
	axis := dominantAxis(f.normal)

	points := make([]vector2.Float64, 0, 3+len(cuts)*2+len(landed))
	source := make([]vector3.Float64, 0, 3+len(cuts)*2+len(landed))
	global := make([]int, 0, 3+len(cuts)*2+len(landed))

	// Every point keeps the 3D coordinates it arrived with. Flattening and
	// lifting a point back is not a round trip, and there is no reason to put
	// one through it when the original is right here.
	indexOf := func(p vector3.Float64, id int) int {
		flat := dropAxis(p, axis)
		for i, existing := range points {
			if existing.Sub(flat).Length() <= tolerance {
				return i
			}
		}
		points = append(points, flat)
		source = append(source, p)
		global = append(global, id)
		return len(points) - 1
	}

	for k, v := range f.verts {
		indexOf(v, ids[k])
	}
	if len(points) != 3 {
		return []face{f}, [][3]int{ids}, nil
	}

	corners := [3]vector2.Float64{points[0], points[1], points[2]}
	edges := [][2]int{{0, 1}, {1, 2}, {2, 0}}

	for _, cut := range cuts {
		from, to := indexOf(cut[0], weld.Index(cut[0])), indexOf(cut[1], weld.Index(cut[1]))
		if from != to {
			edges = append(edges, [2]int{from, to})
		}
	}

	// Carried by no edge of their own; they are here so the triangulator
	// splits whatever edge they sit on.
	for _, p := range landed {
		indexOf(p, weld.Index(p))
	}

	if len(points) == 3 {
		return []face{f}, [][3]int{ids}, nil
	}

	flat, err := triangulation.ConstrainedDelaunayEdges(points, edges)
	if err != nil {
		return nil, nil, fmt.Errorf("cutting a face along %d segments: %w", len(cuts), err)
	}

	indices := flat.Indices()
	positions := flat.Float3Attribute(modeling.PositionAttribute)

	// Input indices survive the triangulation, so anything past what was fed
	// in is a point it invented and the only one needing a coordinate built
	// for it.
	resolve := func(index int) (vector2.Float64, vector3.Float64, int) {
		v := positions.At(index)
		flattened := vector2.New(v.X(), v.Z())
		if index < len(source) {
			return flattened, source[index], global[index]
		}
		lifted := liftOntoPlane(flattened, axis, f)
		return flattened, lifted, weld.Index(lifted)
	}

	pieces := make([]face, 0, indices.Len()/3)
	pieceIDs := make([][3]int, 0, indices.Len()/3)

	for i := 0; i+2 < indices.Len(); i += 3 {
		var flatCorners [3]vector2.Float64
		var solid [3]vector3.Float64
		var corner [3]int
		for k := 0; k < 3; k++ {
			flatCorners[k], solid[k], corner[k] = resolve(indices.At(i + k))
		}

		// A cut endpoint landing a hair outside the triangle would widen the
		// hull the triangulator fills, so anything beyond the original face
		// is dropped rather than carried into the result.
		middle := flatCorners[0].Add(flatCorners[1]).Add(flatCorners[2]).Scale(1. / 3.)
		if !within(corners, middle, tolerance) {
			continue
		}

		// A point a rounding error off an edge makes a triangle with no
		// height, and the face across that edge does not see it.
		if heightless(flatCorners, tolerance) {
			continue
		}

		piece, ok := newFace(solid[0], solid[1], solid[2])
		if !ok {
			continue
		}
		if piece.normal.Dot(f.normal) < 0 {
			piece = piece.reversed()
			corner[1], corner[2] = corner[2], corner[1]
		}
		pieces = append(pieces, piece)
		pieceIDs = append(pieceIDs, corner)
	}

	if len(pieces) == 0 {
		return []face{f}, [][3]int{ids}, nil
	}
	return pieces, pieceIDs, nil
}

func heightless(corners [3]vector2.Float64, tolerance float64) bool {
	longest := 0.
	for i := range corners {
		longest = math.Max(longest, corners[i].Distance(corners[(i+1)%3]))
	}
	return math.Abs(predicate.Orient2D(corners[0], corners[1], corners[2])) <= tolerance*longest
}

func within(corners [3]vector2.Float64, p vector2.Float64, tolerance float64) bool {
	side := func(a, b vector2.Float64) float64 {
		return (b.X()-a.X())*(p.Y()-a.Y()) - (p.X()-a.X())*(b.Y()-a.Y())
	}

	positive, negative := false, false
	for i := 0; i < 3; i++ {
		d := side(corners[i], corners[(i+1)%3])
		if d > tolerance {
			positive = true
		}
		if d < -tolerance {
			negative = true
		}
	}
	return !(positive && negative)
}
