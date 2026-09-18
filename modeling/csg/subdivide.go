package csg

import (
	"errors"
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// The paper shares one vertex array so face adjacency is known. Carrying only
// corners means adjacency has to be recovered by position.
type face struct {
	verts  geometry.Triangle
	normal vector3.Float64

	// The source triangle this was cut from, and each corner as barycentric
	// weights over it, which is all vertex data needs to follow the cut.
	parent  int
	weights [3][3]float64
}

var ownCorners = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

// A triangle enclosing no area contributes no surface and has no normal to
// steer section 7's ray by, so it is dropped rather than carried.
func newFace(verts geometry.Triangle) (face, bool) {
	normal := verts.Normal()
	if normal.Length() == 0 {
		return face{}, false
	}
	return face{verts: verts, normal: normal}, true
}

func (f face) reversed() face {
	return face{
		verts:   geometry.Triangle{f.verts[0], f.verts[2], f.verts[1]},
		normal:  f.normal.Scale(-1),
		parent:  f.parent,
		weights: [3][3]float64{f.weights[0], f.weights[2], f.weights[1]},
	}
}

func (f face) plane() geometry.Plane {
	return geometry.NewPlane(f.verts[0], f.normal)
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
		return nil, fmt.Errorf("mesh is %v, not triangles", m.Topology())
	}
	if !m.HasFloat3Attribute(modeling.PositionAttribute) {
		return nil, errors.New("mesh carries no position attribute")
	}

	indices := m.Indices()
	positions := m.Float3Attribute(modeling.PositionAttribute)

	faces := make([]face, 0, indices.Len()/3)
	for i := 0; i+2 < indices.Len(); i += 3 {
		f, ok := newFace(geometry.Triangle{
			positions.At(indices.At(i)),
			positions.At(indices.At(i + 1)),
			positions.At(indices.At(i + 2)),
		})
		if ok {
			f.parent, f.weights = i/3, ownCorners
			faces = append(faces, f)
		}
	}
	return faces, nil
}

// Tolerance scales with the mesh being cut. Each mesh uses its own span, so a
// mesh CheckClosed accepted alone is accepted here too.
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
	bounds := faces[0].verts.BoundingBox()
	for _, f := range faces[1:] {
		bounds.EncapsulateBounds(f.verts.BoundingBox())
	}
	return bounds.Size().Length()
}

// Section 4: where faces of against cross each face of target. Both solids
// split at every curvePoint. touching marks target faces coplanar with against.
func cutsAgainst(target, against []face, tolerance float64) (cuts [][]segment, curvePoints []vector3.Float64, touching []bool) {
	boxes := make([]trees.Element, len(against))
	for i, f := range against {
		boxes[i] = faceElement{bounds: f.verts.BoundingBox()}
	}
	tree := trees.NewOctree(boxes)

	cuts = make([][]segment, len(target))
	curvePoints = make([]vector3.Float64, 0)
	touching = make([]bool, len(target))

	for i, f := range target {
		reach := f.verts.Reach()
		for _, j := range tree.ElementsWithinRange(f.verts.Centroid(), reach+tolerance) {
			other := against[j]
			shared, ok := f.verts.Intersect(other.verts, tolerance)
			if !ok {
				if other.plane().Holds(f.verts, tolerance) {
					touching[i] = true
				}
				continue
			}
			cut := segment{shared.GetStartPoint(), shared.GetEndPoint()}
			cuts[i] = append(cuts[i], cut)
			curvePoints = append(curvePoints, cut[0], cut[1])
		}
	}

	return cuts, curvePoints, touching
}

// Splits every face along its cuts and at any curve point landing on it.
// onCurve collects the ids the pieces carry at those points.
func splitAll(
	target []face,
	cornerIDs [][3]int,
	cuts [][]segment,
	curvePoints []vector3.Float64,
	touching []bool,
	tolerance float64,
	ids *idSpace,
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
		reach := f.verts.Reach()
		landedPoints := make([]vector3.Float64, 0)
		for _, j := range tree.ElementsWithinRange(f.verts.Centroid(), reach+tolerance) {
			if f.verts.Contains(curvePoints[j], tolerance) {
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

		pieces, pieceIDs, err := splitFace(f, cornerIDs[i], cuts[i], landedPoints, tolerance, ids, onCurve)
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

// Section 6. The paper subdivides case by case to keep pieces convex; here
// the cuts go to a constrained triangulator as edges it must keep.
func splitFace(
	f face,
	cornerIDs [3]int,
	cuts []segment,
	landedPoints []vector3.Float64,
	tolerance float64,
	ids *idSpace,
	onCurve map[int]bool,
) ([]face, [][3]int, error) {
	axis := geometry.DominantAxis(f.normal)

	// One entry per triangulator point: its plane position, space position,
	// welded id, and whether it came from a cut.
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
		flat = append(flat, geometry.DropAxis(point, axis))
		world = append(world, point)
		weldedIDs = append(weldedIDs, weldedID)
		onCut = append(onCut, false)
		return len(flat) - 1
	}
	place := func(point vector3.Float64) int {
		point = snapToEdge(f, point, tolerance)
		i, known := indexOf(point)
		if !known {
			i = record(point, ids.index(point))
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

	outline := geometry.Triangle2D{flat[0], flat[1], flat[2]}
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
		lifted := f.plane().Lift(inPlane, axis)
		weldedID := ids.index(lifted)
		onCurve[weldedID] = true
		return inPlane, lifted, weldedID
	}

	pieces := make([]face, 0, indices.Len()/3)
	pieceIDs := make([][3]int, 0, indices.Len()/3)

	for i := 0; i+2 < indices.Len(); i += 3 {
		var flatCorners geometry.Triangle2D
		var worldCorners geometry.Triangle
		var pieceCornerIDs [3]int
		for k := 0; k < 3; k++ {
			flatCorners[k], worldCorners[k], pieceCornerIDs[k] = pointAt(indices.At(i + k))
		}

		// A cut endpoint a hair outside the triangle would widen the hull the
		// triangulator fills, so anything beyond the original face is dropped.
		centroid := flatCorners[0].Add(flatCorners[1]).Add(flatCorners[2]).Scale(1. / 3.)
		if !outline.Contains(centroid, tolerance) {
			continue
		}

		// A point on the outline leaves the triangulator a zero-area triangle
		// along that edge, which the face across the edge never sees.
		if worldCorners.Degenerate(tolerance) {
			continue
		}

		piece, ok := newFace(worldCorners)
		if !ok {
			continue
		}
		// Barycentric coordinates survive the axis drop, the projection being
		// linear and the outline non-degenerate.
		piece.parent = f.parent
		for k := 0; k < 3; k++ {
			piece.weights[k] = f.blend(outline.Barycentric(flatCorners[k]))
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
	if onEdge, distance := f.verts.ClosestPointOnEdges(point); distance <= tolerance {
		return onEdge
	}
	return point
}
