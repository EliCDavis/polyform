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
	weights [3]vector3.Float64
}

var ownCorners = [3]vector3.Float64{
	vector3.New(1., 0., 0.),
	vector3.New(0., 1., 0.),
	vector3.New(0., 0., 1.),
}

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
		weights: [3]vector3.Float64{f.weights[0], f.weights[2], f.weights[1]},
	}
}

func (f face) plane() geometry.Plane {
	return geometry.NewPlane(f.verts[0], f.normal)
}

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
func cutsAgainst(target, against []face, tolerance float64) (cuts [][]geometry.Line3D, curvePoints []vector3.Float64, touching []bool) {
	boxes := make([]trees.Element, len(against))
	for i, f := range against {
		boxes[i] = faceElement{bounds: f.verts.BoundingBox()}
	}
	tree := trees.NewOctree(boxes)

	cuts = make([][]geometry.Line3D, len(target))
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
					for _, cut := range coplanarCuts(f, other, tolerance) {
						cuts[i] = append(cuts[i], cut)
						curvePoints = append(curvePoints, cut.GetStartPoint(), cut.GetEndPoint())
					}
				}
				continue
			}
			cuts[i] = append(cuts[i], shared)
			curvePoints = append(curvePoints, shared.GetStartPoint(), shared.GetEndPoint())
		}
	}

	return cuts, curvePoints, touching
}

// Points one face's snap merges must merge in every face of both solids, or the seams disagree.
func mergeCurvePoints(target, against []face, cutsTarget, cutsAgainst [][]geometry.Line3D, curvePoints []vector3.Float64, tolerance float64) {
	if len(curvePoints) < 2 {
		return
	}

	indexOf := make(map[vector3.Float64]int, len(curvePoints))
	points := make([]vector3.Float64, 0, len(curvePoints))
	for _, p := range curvePoints {
		if _, ok := indexOf[p]; !ok {
			indexOf[p] = len(points)
			points = append(points, p)
		}
	}

	group := make([]int, len(points))
	for i := range group {
		group[i] = i
	}
	var root func(int) int
	root = func(i int) int {
		if group[i] != i {
			group[i] = root(group[i])
		}
		return group[i]
	}
	join := func(i, j int) {
		ri, rj := root(i), root(j)
		if ri < rj {
			group[rj] = ri
		} else if rj < ri {
			group[ri] = rj
		}
	}

	boxes := make([]trees.Element, len(points))
	for i, p := range points {
		boxes[i] = faceElement{bounds: geometry.NewAABBFromPoints(p)}
	}
	tree := trees.NewOctree(boxes)
	for i, p := range points {
		for _, j := range tree.ElementsWithinRange(p, tolerance) {
			if j != i && points[j].Distance(p) <= tolerance {
				join(i, j)
			}
		}
	}

	edgesOf := make([][]geometry.Line3D, len(points))
	snapGroups := func(faces []face, cuts [][]geometry.Line3D) {
		for i, f := range faces {
			if len(cuts[i]) == 0 {
				continue
			}
			members := make([]int, 0)
			for _, j := range tree.ElementsWithinRange(f.verts.Centroid(), f.verts.Reach()+tolerance) {
				if f.verts.Contains(points[j], tolerance) {
					members = append(members, j)
				}
			}
			if len(members) < 2 {
				continue
			}
			snapped := make([]vector3.Float64, len(members))
			for k, m := range members {
				snapped[k] = snapToEdge(f, points[m], tolerance)
				for e := 0; e < 3; e++ {
					edge := geometry.NewLine3D(f.verts[e], f.verts[(e+1)%3])
					if edge.ClosestPointOnLine(points[m]).Distance(points[m]) <= tolerance {
						edgesOf[m] = append(edgesOf[m], edge)
					}
				}
			}
			for k := range members {
				for l := k + 1; l < len(members); l++ {
					if snapped[k].Distance(snapped[l]) <= tolerance {
						join(members[k], members[l])
					}
				}
			}
		}
	}
	snapGroups(target, cutsTarget)
	snapGroups(against, cutsAgainst)

	members := make(map[int][]int)
	for i := range points {
		r := root(i)
		members[r] = append(members[r], i)
	}
	chosen := make([]vector3.Float64, len(points))
	for _, group := range members {
		edges := make([]geometry.Line3D, 0)
		for _, m := range group {
			edges = append(edges, edgesOf[m]...)
		}
		best, bestOn := group[0], -1
		for _, m := range group {
			on := 0
			for _, edge := range edges {
				if edge.ClosestPointOnLine(points[m]).Distance(points[m]) <= tolerance {
					on++
				}
			}
			if on > bestOn {
				best, bestOn = m, on
			}
		}
		for _, m := range group {
			chosen[m] = points[best]
		}
	}

	moved := func(p vector3.Float64) vector3.Float64 { return chosen[indexOf[p]] }
	for _, cuts := range [][][]geometry.Line3D{cutsTarget, cutsAgainst} {
		for i := range cuts {
			for k, cut := range cuts[i] {
				cuts[i][k] = geometry.NewLine3D(moved(cut.GetStartPoint()), moved(cut.GetEndPoint()))
			}
		}
	}
	for i, p := range curvePoints {
		curvePoints[i] = moved(p)
	}
}

// An edge running along f's outline would cross the outline constraints, so it gives only its endpoints.
func coplanarCuts(f, other face, tolerance float64) []geometry.Line3D {
	cuts := make([]geometry.Line3D, 0, 3)
	for k := 0; k < 3; k++ {
		a := snapToFace(f, other.verts[k], tolerance)
		b := snapToFace(f, other.verts[(k+1)%3], tolerance)
		from, to, ok := clipToFace(f, a, b, tolerance*1e-3)
		if !ok {
			continue
		}
		if from.Distance(to) <= tolerance || alongOutline(f, from, to, tolerance) {
			cuts = append(cuts, geometry.NewLine3D(from, from), geometry.NewLine3D(to, to))
			continue
		}
		cuts = append(cuts, geometry.NewLine3D(from, to))
	}
	return cuts
}

func snapToFace(f face, point vector3.Float64, tolerance float64) vector3.Float64 {
	for _, corner := range f.verts {
		if corner.Distance(point) <= tolerance {
			return corner
		}
	}
	return snapToEdge(f, point, tolerance)
}

func clipToFace(f face, a, b vector3.Float64, slack float64) (vector3.Float64, vector3.Float64, bool) {
	low, high := 0., 1.
	direction := b.Sub(a)
	for i := 0; i < 3; i++ {
		inward := f.normal.Cross(f.verts[(i+1)%3].Sub(f.verts[i])).Normalized()
		start := a.Sub(f.verts[i]).Dot(inward) + slack
		rate := direction.Dot(inward)
		switch {
		case rate == 0:
			if start < 0 {
				return a, b, false
			}
		case rate > 0:
			low = max(low, -start/rate)
		default:
			high = min(high, -start/rate)
		}
	}
	if high < low {
		return a, b, false
	}
	return a.Add(direction.Scale(low)), a.Add(direction.Scale(high)), true
}

func alongOutline(f face, from, to vector3.Float64, tolerance float64) bool {
	for i := 0; i < 3; i++ {
		edge := geometry.NewLine3D(f.verts[i], f.verts[(i+1)%3])
		if edge.ClosestPointOnLine(from).Distance(from) <= tolerance && edge.ClosestPointOnLine(to).Distance(to) <= tolerance {
			return true
		}
	}
	return false
}

// Splits every face along its cuts and at any curve point landing on it.
// onCurve collects the ids the pieces carry at those points.
func splitAll(
	target []face,
	cornerIDs [][3]int,
	cuts [][]geometry.Line3D,
	curvePoints []vector3.Float64,
	touching []bool,
	tolerance float64,
	ids *idSpace,
) (half, error) {
	onCurve := make(map[int]bool, len(curvePoints))
	offCurve := make(map[[2]int]bool)
	if len(curvePoints) == 0 {
		return half{target, cornerIDs, touching, onCurve, offCurve}, nil
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
		offCurve:  offCurve,
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

		pieces, pieceIDs, err := splitFace(f, cornerIDs[i], cuts[i], landedPoints, tolerance, ids, onCurve, offCurve)
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
	cuts []geometry.Line3D,
	landedPoints []vector3.Float64,
	tolerance float64,
	ids *idSpace,
	onCurve map[int]bool,
	offCurve map[[2]int]bool,
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

	cutLines := make([]geometry.Line3D, 0, len(cuts))
	for _, cut := range cuts {
		from, to := place(cut.GetStartPoint()), place(cut.GetEndPoint())
		if from != to {
			constraints = append(constraints, [2]int{from, to})
			cutLines = append(cutLines, geometry.NewLine3D(world[from], world[to]))
		}
	}
	alongCut := func(a, b vector3.Float64) bool {
		for _, line := range cutLines {
			if line.ClosestPointOnLine(a).Distance(a) <= tolerance && line.ClosestPointOnLine(b).Distance(b) <= tolerance {
				return true
			}
		}
		return false
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
		if worldCorners.Degenerate(tolerance) && onOneEdge(f, worldCorners, tolerance) {
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

		for k := 0; k < 3; k++ {
			a, b := piece.verts[k], piece.verts[(k+1)%3]
			if alongCut(a, b) || alongOutline(f, a, b, tolerance) {
				continue
			}
			from, to := pieceCornerIDs[k], pieceCornerIDs[(k+1)%3]
			offCurve[[2]int{min(from, to), max(from, to)}] = true
		}
	}

	if len(pieces) == 0 {
		return []face{f}, [][3]int{cornerIDs}, nil
	}
	return pieces, pieceIDs, nil
}

func onOneEdge(f face, corners geometry.Triangle, tolerance float64) bool {
	for i := 0; i < 3; i++ {
		edge := geometry.NewLine3D(f.verts[i], f.verts[(i+1)%3])
		if edge.ClosestPointOnLine(corners[0]).Distance(corners[0]) <= tolerance &&
			edge.ClosestPointOnLine(corners[1]).Distance(corners[1]) <= tolerance &&
			edge.ClosestPointOnLine(corners[2]).Distance(corners[2]) <= tolerance {
			return true
		}
	}
	return false
}

// Weights over this face's corners re-expressed over its parent's.
func (f face) blend(local vector3.Float64) vector3.Float64 {
	return f.weights[0].Scale(local.X()).
		Add(f.weights[1].Scale(local.Y())).
		Add(f.weights[2].Scale(local.Z()))
}

// A point within tolerance of an edge is moved onto it, so the face across
// that edge sees the same point and no sliver forms between the two.
func snapToEdge(f face, point vector3.Float64, tolerance float64) vector3.Float64 {
	if onEdge, distance := f.verts.ClosestPointOnEdges(point); distance <= tolerance {
		return onEdge
	}
	return point
}
