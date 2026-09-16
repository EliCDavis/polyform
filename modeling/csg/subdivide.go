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

	// The triangle of the source mesh this was cut from, and each corner as
	// barycentric weights over that triangle's corners, which is all vertex
	// data needs to follow the cut.
	parent  int
	weights [3][3]float64
}

var ownCorners = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

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
		verts:   [3]vector3.Float64{f.verts[0], f.verts[2], f.verts[1]},
		normal:  f.normal.Scale(-1),
		parent:  f.parent,
		weights: [3][3]float64{f.weights[0], f.weights[2], f.weights[1]},
	}
}

// Section 7 calls the average of a polygon's vertices its barycenter.
func (f face) barycenter() vector3.Float64 {
	return f.verts[0].
		Add(f.verts[1]).
		Add(f.verts[2]).
		Scale(1. / 3.)
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
	return max(
		f.verts[0].Sub(center).Length(),
		f.verts[1].Sub(center).Length(),
		f.verts[2].Sub(center).Length(),
	)
}

// Which side of other's plane each corner falls on. Not a distance: the
// magnitude carries other's area, which cancels in the ratios callers take.
func (f face) sidesOf(other face) [3]float64 {
	var sides [3]float64
	for i, corner := range f.verts {
		sides[i] = predicate.Orient3D(other.verts[0], other.verts[1], other.verts[2], corner)
	}
	return sides
}

type segment [2]vector3.Float64

type faceElement struct {
	bounds geometry.AABB
}

func (e faceElement) BoundingBox() geometry.AABB { return e.bounds }

func (e faceElement) ClosestPoint(point vector3.Float64) vector3.Float64 {
	return e.bounds.ClosestPoint(point)
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
			f.parent, f.weights = i/3, ownCorners
			faces = append(faces, f)
		}
	}
	return faces, nil
}

// Every comparison against zero here is really a comparison against the size
// of what is being cut, so a millimetre model and a kilometre one behave the
// same. Each mesh's own span, never the pair's: a mesh CheckClosed accepted
// alone must be accepted here, however far from its partner it sits.
func toleranceFor(first, second []face) float64 {
	span := max(spanOf(first), spanOf(second))
	if span == 0 || math.IsInf(span, 0) || math.IsNaN(span) {
		return 1e-9
	}
	return span * 1e-9
}

func spanOf(faces []face) float64 {
	if len(faces) == 0 {
		return 0
	}
	bounds := faces[0].bounds()
	for _, f := range faces[1:] {
		bounds.EncapsulateBounds(f.bounds())
	}
	return bounds.Size().Length()
}

// Section 5, "Do Two Polygons Intersect?": the segment shared by two faces.
//
// Reports false for coplanar pairs. Section 4's outer loop leaves those
// alone, expecting the faces around them to do the cutting.
func sharedSegment(f, other face, tolerance float64) (segment, bool) {
	faceSides := f.sidesOf(other)
	otherSides := other.sidesOf(f)

	if entirelyOneSide(faceSides) || entirelyOneSide(otherSides) {
		return segment{}, false
	}
	// Within tolerance counts as coplanar: any closer and the cut direction
	// below is a cross product made of rounding error.
	if sharesPlane(f, other, tolerance) || sharesPlane(other, f, tolerance) {
		return segment{}, false
	}

	faceCrossing, ok := crossesPlane(f, faceSides, tolerance)
	if !ok {
		return segment{}, false
	}
	otherCrossing, ok := crossesPlane(other, otherSides, tolerance)
	if !ok {
		return segment{}, false
	}

	// Both segments lie on the line where the two planes meet, so they can be
	// compared as intervals along it.
	direction := f.normal.Cross(other.normal)
	if direction.Length() == 0 {
		return segment{}, false
	}
	direction = direction.Normalized()

	base := faceCrossing[0]
	along := func(point vector3.Float64) float64 { return point.Sub(base).Dot(direction) }

	// The overlap always ends on one of the four crossing points already in
	// hand, so the answer is which one, not where. Rebuilding it from a
	// parameter would round it through a normalized direction and a scale,
	// and the same point reached from the other face would land elsewhere -
	// which is the drift the snapping tolerances downstream exist to absorb.
	orderedAlong := func(crossing [2]vector3.Float64) (start, end vector3.Float64, startAt, endAt float64) {
		first, second := along(crossing[0]), along(crossing[1])
		if first <= second {
			return crossing[0], crossing[1], first, second
		}
		return crossing[1], crossing[0], second, first
	}

	faceStart, faceEnd, faceStartAt, faceEndAt := orderedAlong(faceCrossing)
	otherStart, otherEnd, otherStartAt, otherEndAt := orderedAlong(otherCrossing)

	start, startAt := faceStart, faceStartAt
	if otherStartAt > startAt {
		start, startAt = otherStart, otherStartAt
	}

	end, endAt := faceEnd, faceEndAt
	if otherEndAt < endAt {
		end, endAt = otherEnd, otherEndAt
	}

	if endAt-startAt <= tolerance {
		return segment{}, false
	}

	return segment{start, end}, true
}

// Against zero, not a tolerance: Orient3D is exact in sign.
func entirelyOneSide(sides [3]float64) bool {
	positive, negative := 0, 0
	for _, side := range sides {
		if side > 0 {
			positive++
		}
		if side < 0 {
			negative++
		}
	}
	return positive == 3 || negative == 3
}

// Every corner of f within tolerance of other's plane.
func sharesPlane(f, other face, tolerance float64) bool {
	for _, corner := range f.verts {
		if math.Abs(other.planeDistance(corner)) > tolerance {
			return false
		}
	}
	return true
}

// Where a face meets a plane: two points, or nothing when it only grazes a
// single corner. sides is which side of that plane each corner falls on.
func crossesPlane(f face, sides [3]float64, tolerance float64) ([2]vector3.Float64, bool) {
	crossings := make([]vector3.Float64, 0, 2)

	record := func(point vector3.Float64) {
		for _, existing := range crossings {
			if existing.Sub(point).Length() <= tolerance {
				return
			}
		}
		crossings = append(crossings, point)
	}

	for i := 0; i < 3; i++ {
		j := (i + 1) % 3
		startSide, endSide := sides[i], sides[j]

		if startSide == 0 {
			record(f.verts[i])
			continue
		}
		if endSide == 0 || (startSide > 0) == (endSide > 0) {
			continue
		}

		// The area factor in both cancels, leaving the true fraction.
		fraction := startSide / (startSide - endSide)
		record(f.verts[i].Add(f.verts[j].Sub(f.verts[i]).Scale(fraction)))
	}

	if len(crossings) != 2 {
		return [2]vector3.Float64{}, false
	}
	return [2]vector3.Float64{crossings[0], crossings[1]}, true
}

// Section 4, "Intersecting the Objects": where the faces of against cross
// each face of target.
//
// The endpoints come back as curvePoints because both solids have to be split
// at all of them, not just at their own. The curve where the two meet belongs
// to both, and a point only one of them splits at leaves the other with an
// edge running straight past it.
// touching reports the target faces that share a plane with a face of the
// other solid. Section 4 leaves such pairs unsplit, so no cut ever separates
// them from their neighbours even though they classify differently.
func cutsAgainst(target, against []face, tolerance float64) (cuts [][]segment, curvePoints []vector3.Float64, touching []bool) {
	boxes := make([]trees.Element, len(against))
	for i, f := range against {
		boxes[i] = faceElement{bounds: f.bounds()}
	}
	tree := trees.NewOctree(boxes)

	cuts = make([][]segment, len(target))
	curvePoints = make([]vector3.Float64, 0)
	touching = make([]bool, len(target))

	for i, f := range target {
		reach := f.reach()
		for _, j := range tree.ElementsWithinRange(f.barycenter(), reach+tolerance) {
			other := against[j]
			cut, ok := sharedSegment(f, other, tolerance)
			if !ok {
				if sharesPlane(f, other, tolerance) {
					touching[i] = true
				}
				continue
			}
			cuts[i] = append(cuts[i], cut)
			curvePoints = append(curvePoints, cut[0], cut[1])
		}
	}

	return cuts, curvePoints, touching
}

// Splits every face along its own cuts, and at any curve point from either
// solid that happens to land on it. onCurve collects the ids the pieces carry
// at those points, so the curve is known by id and never by position.
func splitAll(
	target []face,
	cornerIDs [][3]int,
	cuts [][]segment,
	curvePoints []vector3.Float64,
	touching []bool,
	tolerance float64,
	weld *welder,
) (half, error) {
	onCurve := make(map[int]bool, len(curvePoints))
	if len(curvePoints) == 0 {
		return half{target, cornerIDs, touching, onCurve}, nil
	}

	boxes := make([]trees.Element, len(curvePoints))
	for i, point := range curvePoints {
		boxes[i] = faceElement{bounds: geometry.NewAABBFromPoints(point)}
	}
	tree := trees.NewOctree(boxes)

	out := half{
		faces:     make([]face, 0, len(target)),
		cornerIDs: make([][3]int, 0, len(target)),
		touching:  make([]bool, 0, len(target)),
		onCurve:   onCurve,
	}

	for i, f := range target {
		reach := f.reach()
		landedPoints := make([]vector3.Float64, 0)
		for _, j := range tree.ElementsWithinRange(f.barycenter(), reach+tolerance) {
			if onFace(f, curvePoints[j], tolerance) {
				landedPoints = append(landedPoints, curvePoints[j])
			}
		}

		// An uncut face keeps the ids it came in with, which is most of them.
		if len(cuts[i]) == 0 && len(landedPoints) == 0 {
			out.faces = append(out.faces, f)
			out.cornerIDs = append(out.cornerIDs, cornerIDs[i])
			out.touching = append(out.touching, touching[i])
			continue
		}

		pieces, pieceIDs, err := splitFace(f, cornerIDs[i], cuts[i], landedPoints, tolerance, weld, onCurve)
		if err != nil {
			return half{}, err
		}
		out.faces = append(out.faces, pieces...)
		out.cornerIDs = append(out.cornerIDs, pieceIDs...)
		for range pieces {
			out.touching = append(out.touching, touching[i])
		}
	}

	return out, nil
}

func onFace(f face, point vector3.Float64, tolerance float64) bool {
	if math.Abs(point.Sub(f.verts[0]).Dot(f.normal)) > tolerance {
		return false
	}
	for i := 0; i < 3; i++ {
		j := (i + 1) % 3
		edge := f.verts[j].Sub(f.verts[i])
		if edge.Cross(point.Sub(f.verts[i])).Dot(f.normal) < -tolerance*edge.Length() {
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
	cornerIDs [3]int,
	cuts []segment,
	landedPoints []vector3.Float64,
	tolerance float64,
	weld *welder,
	onCurve map[int]bool,
) ([]face, [][3]int, error) {
	axis := dominantAxis(f.normal)

	// One entry per point handed to the triangulator: where it sits in the
	// plane, where it sits in space, the welded id it carries, and whether it
	// came from a cut.
	capacity := 3 + len(cuts)*2 + len(landedPoints)
	flat := make([]vector2.Float64, 0, capacity)
	world := make([]vector3.Float64, 0, capacity)
	weldedIDs := make([]int, 0, capacity)
	onCut := make([]bool, 0, capacity)

	// Identity is decided in 3D, the way the welder decides it, so a point
	// that is a corner here is the same corner in every face that has it.
	indexOf := func(point vector3.Float64) (int, bool) {
		for i, existing := range world {
			if existing.Distance(point) <= tolerance {
				return i, true
			}
		}
		return -1, false
	}
	record := func(point vector3.Float64, weldedID int) int {
		flat = append(flat, dropAxis(point, axis))
		world = append(world, point)
		weldedIDs = append(weldedIDs, weldedID)
		onCut = append(onCut, false)
		return len(flat) - 1
	}
	place := func(point vector3.Float64) int {
		point = snapToEdge(f, point, tolerance)
		i, known := indexOf(point)
		if !known {
			i = record(point, weld.Index(point))
		}
		onCut[i] = true
		return i
	}

	for k, corner := range f.verts {
		record(corner, cornerIDs[k])
	}
	if len(flat) != 3 {
		return []face{f}, [][3]int{cornerIDs}, nil
	}

	outline := [3]vector2.Float64{flat[0], flat[1], flat[2]}
	constraints := [][2]int{{0, 1}, {1, 2}, {2, 0}}

	for _, cut := range cuts {
		from, to := place(cut[0]), place(cut[1])
		if from != to {
			constraints = append(constraints, [2]int{from, to})
		}
	}

	// Carried by no edge of their own; they are here so the triangulator
	// splits whatever edge they sit on.
	for _, point := range landedPoints {
		place(point)
	}

	for i, fromCut := range onCut {
		if fromCut {
			onCurve[weldedIDs[i]] = true
		}
	}

	if len(flat) == 3 {
		return []face{f}, [][3]int{cornerIDs}, nil
	}

	triangulated, err := triangulation.ConstrainedDelaunayEdges(flat, constraints)
	if err != nil {
		return nil, nil, fmt.Errorf("cutting a face along %d segments: %w", len(cuts), err)
	}

	indices := triangulated.Indices()
	positions := triangulated.Float3Attribute(modeling.PositionAttribute)

	// Input indices survive the triangulation, so anything past what was fed
	// in is a point it invented where two cuts cross, and lies on the curve.
	pointAt := func(index int) (vector2.Float64, vector3.Float64, int) {
		position := positions.At(index)
		inPlane := vector2.New(position.X(), position.Z())
		if index < len(world) {
			return inPlane, world[index], weldedIDs[index]
		}
		lifted := liftOntoPlane(inPlane, axis, f)
		weldedID := weld.Index(lifted)
		onCurve[weldedID] = true
		return inPlane, lifted, weldedID
	}

	pieces := make([]face, 0, indices.Len()/3)
	pieceIDs := make([][3]int, 0, indices.Len()/3)

	for i := 0; i+2 < indices.Len(); i += 3 {
		var flatCorners [3]vector2.Float64
		var worldCorners [3]vector3.Float64
		var pieceCornerIDs [3]int
		for k := 0; k < 3; k++ {
			flatCorners[k], worldCorners[k], pieceCornerIDs[k] = pointAt(indices.At(i + k))
		}

		// A cut endpoint landing a hair outside the triangle would widen the
		// hull the triangulator fills, so anything beyond the original face
		// is dropped rather than carried into the result.
		centroid := flatCorners[0].Add(flatCorners[1]).Add(flatCorners[2]).Scale(1. / 3.)
		if !withinOutline(outline, centroid, tolerance) {
			continue
		}

		// A point on the outline leaves the triangulator a zero-area triangle
		// along that edge, which the face across the edge never sees.
		if heightless(worldCorners, tolerance) {
			continue
		}

		piece, ok := newFace(worldCorners[0], worldCorners[1], worldCorners[2])
		if !ok {
			continue
		}
		piece.parent = f.parent
		for k := 0; k < 3; k++ {
			piece.weights[k] = f.blend(barycentric(outline, flatCorners[k]))
		}
		if piece.normal.Dot(f.normal) < 0 {
			piece = piece.reversed()
			pieceCornerIDs[1], pieceCornerIDs[2] = pieceCornerIDs[2], pieceCornerIDs[1]
		}
		pieces = append(pieces, piece)
		pieceIDs = append(pieceIDs, pieceCornerIDs)
	}

	if len(pieces) == 0 {
		return []face{f}, [][3]int{cornerIDs}, nil
	}
	return pieces, pieceIDs, nil
}

// Barycentric coordinates survive the axis drop, the projection being linear
// and the outline non-degenerate, so they are read off the flattened points.
func barycentric(outline [3]vector2.Float64, point vector2.Float64) [3]float64 {
	twiceArea := func(a, b, c vector2.Float64) float64 {
		return (b.X()-a.X())*(c.Y()-a.Y()) - (c.X()-a.X())*(b.Y()-a.Y())
	}
	whole := twiceArea(outline[0], outline[1], outline[2])
	return [3]float64{
		twiceArea(point, outline[1], outline[2]) / whole,
		twiceArea(outline[0], point, outline[2]) / whole,
		twiceArea(outline[0], outline[1], point) / whole,
	}
}

// Weights over this face's corners re-expressed over its parent's.
func (f face) blend(local [3]float64) [3]float64 {
	var out [3]float64
	for k := 0; k < 3; k++ {
		for j := 0; j < 3; j++ {
			out[j] += local[k] * f.weights[k][j]
		}
	}
	return out
}

// A point within tolerance of an edge is moved onto it, so the face across
// that edge sees the same point and no sliver forms between the two.
func snapToEdge(f face, point vector3.Float64, tolerance float64) vector3.Float64 {
	nearest, nearestDistance := point, tolerance
	for i := 0; i < 3; i++ {
		onEdge := geometry.NewLine3D(f.verts[i], f.verts[(i+1)%3]).ClosestPointOnLine(point)
		if distance := onEdge.Distance(point); distance <= nearestDistance {
			nearest, nearestDistance = onEdge, distance
		}
	}
	return nearest
}

func heightless(corners [3]vector3.Float64, tolerance float64) bool {
	longest := 0.
	for i := range corners {
		longest = max(longest, corners[i].Distance(corners[(i+1)%3]))
	}
	twiceArea := corners[1].Sub(corners[0]).Cross(corners[2].Sub(corners[0])).Length()
	return twiceArea <= tolerance*longest
}

// Orient2D carries the edge's length, so the band is tolerance in distance.
func withinOutline(outline [3]vector2.Float64, point vector2.Float64, tolerance float64) bool {
	positive, negative := false, false
	for i := 0; i < 3; i++ {
		start, end := outline[i], outline[(i+1)%3]
		side := predicate.Orient2D(start, end, point)
		band := tolerance * start.Distance(end)
		if side > band {
			positive = true
		}
		if side < -band {
			negative = true
		}
	}
	return !(positive && negative)
}
