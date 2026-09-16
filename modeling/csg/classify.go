package csg

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type solid struct {
	faces     []face
	tree      *trees.OctTree
	bounds    geometry.AABB
	tolerance float64
}

// Number of faces heuristic for when we should swap over to an octree.
const scanBudget = 2_000_000

func newSolid(faces []face, tolerance float64, queries int) *solid {
	worthIndexing := queries*len(faces) >= scanBudget

	var boxes []trees.Element
	if worthIndexing {
		boxes = make([]trees.Element, len(faces))
	}

	// Grown from the face boxes already being built for the tree, rather
	// than from a second list holding every vertex in the mesh.
	var bounds geometry.AABB
	for i, f := range faces {
		box := f.bounds()
		if worthIndexing {
			boxes[i] = faceElement{bounds: box}
		}
		if i == 0 {
			bounds = box
			continue
		}
		bounds.EncapsulateBounds(box)
	}

	s := &solid{
		faces:     faces,
		bounds:    bounds,
		tolerance: tolerance,
	}
	if worthIndexing {
		s.tree = trees.NewOctree(boxes)
	}
	return s
}

// The tree only ever narrows the field; callers re-test whatever it hands
// back, so scanning everything reaches the same answer by a slower road.
// Visiting stops early when visit returns false.
func (s *solid) eachNear(origin vector3.Float64, radius float64, visit func(faceIndex int) bool) {
	if s.tree == nil {
		s.eachFace(visit)
		return
	}
	for _, faceIndex := range s.tree.ElementsWithinRange(origin, radius) {
		if !visit(faceIndex) {
			return
		}
	}
}

func (s *solid) eachAlong(ray geometry.Ray, min, max float64, visit func(faceIndex int) bool) {
	if s.tree == nil {
		s.eachFace(visit)
		return
	}
	for _, faceIndex := range s.tree.ElementsIntersectingRay(ray, min, max) {
		if !visit(faceIndex) {
			return
		}
	}
}

func (s *solid) eachFace(visit func(faceIndex int) bool) {
	for faceIndex := range s.faces {
		if !visit(faceIndex) {
			return
		}
	}
}

// How far outside a triangle a hit may land and still count. A ray leaving
// through an edge two triangles share sits a rounding error to one side of
// both of them, and rejecting it from each in turn loses the crossing
// altogether: the ray then appears to leave the solid without touching it.
const rim = 1e-9

// Möller-Trumbore. margin reports how close the hit landed to the triangle's
// rim, in barycentric terms, which is what says whether it can be trusted.
func (f face) rayHit(origin, direction vector3.Float64) (distance, margin float64, hit bool) {
	edge1 := f.verts[1].Sub(f.verts[0])
	edge2 := f.verts[2].Sub(f.verts[0])

	perpendicular := direction.Cross(edge2)
	determinant := edge1.Dot(perpendicular)
	if determinant == 0 {
		return 0, 0, false
	}

	inverse := 1 / determinant
	fromCorner := origin.Sub(f.verts[0])
	u := fromCorner.Dot(perpendicular) * inverse
	crossed := fromCorner.Cross(edge1)
	v := direction.Dot(crossed) * inverse

	if u < -rim || v < -rim || u+v > 1+rim {
		return 0, 0, false
	}

	return edge2.Dot(crossed) * inverse, min(u, v, 1-u-v), true
}

func (f face) planeDistance(point vector3.Float64) float64 {
	return point.Sub(f.verts[0]).Dot(f.normal)
}

// Section 7, "Classifying Polygons".
//
// "A ray is cast from the barycenter of polygonA in the direction of the
// normal vector to polygonA, and is intersected with every polygonB in
// objectB. The polygonB that intersects the ray closest to the barycenter is
// found."
//
// The paper stops there. A barycenter ray fired along a face normal lands on
// a shared edge often enough in practice that a hit too near a triangle's rim
// is retried along a nudged direction, which the paper leaves open.
func (s *solid) classify(f face) classification {
	origin := f.barycenter()

	verdict, liesOnSurface, touchesSurface := s.onBoundary(f, origin)
	if liesOnSurface {
		return verdict
	}
	// A barycenter on the other surface without the face lying in it: a ray
	// from there begins on a boundary and settles nothing.
	if touchesSurface {
		return s.enclosed(origin)
	}

	for _, direction := range nudged(f.normal) {
		if verdict, settled := s.castFrom(origin, direction); settled {
			return verdict
		}
	}
	return s.enclosed(origin)
}

// Every ray grazed a rim. The winding number has no direction to graze
// along: for a closed solid it is a whole turn inside and none outside.
func (s *solid) enclosed(origin vector3.Float64) classification {
	turns := 0.
	for _, f := range s.faces {
		turns += geometry.SolidAngle(
			f.verts[0].Sub(origin),
			f.verts[1].Sub(origin),
			f.verts[2].Sub(origin),
		)
	}
	if math.Abs(turns) > 2*math.Pi {
		return inside
	}
	return outside
}

// "If the origin of the ray lies in the plane of the nearest polygonB, then
// polygonA lies in the boundary of objectB. In this case, if the normal
// vectors of polygonA and polygonB point in the same direction polygonA is
// classified as SAME; otherwise, it is classified as OPPOSITE."
//
// The whole face has to lie in that plane, not just its barycenter.
// touchesSurface reports a barycenter on the surface without the face on it.
func (s *solid) onBoundary(f face, origin vector3.Float64) (verdict classification, liesOnSurface, touchesSurface bool) {
	s.eachNear(origin, s.tolerance, func(faceIndex int) bool {
		other := s.faces[faceIndex]
		if math.Abs(other.planeDistance(origin)) > s.tolerance {
			return true
		}
		if _, _, hit := other.rayHit(origin, other.normal); !hit {
			return true
		}
		if !sharesPlane(f, other, s.tolerance) {
			touchesSurface = true
			return true
		}
		liesOnSurface = true
		if f.normal.Dot(other.normal) > 0 {
			verdict = same
		} else {
			verdict = opposite
		}
		return false
	})
	return verdict, liesOnSurface, touchesSurface
}

// "If the normal to polygonB points toward polygonA, then polygonA is OUTSIDE
// objectB; otherwise, polygonA is INSIDE objectB. If no polygons were
// intersected, then polygonA is OUTSIDE objectB."
//
// settled is false when the ray grazed a rim and its answer cannot be trusted.
func (s *solid) castFrom(origin, direction vector3.Float64) (verdict classification, settled bool) {
	reach := s.bounds.Size().Length() * 2
	if reach == 0 {
		return outside, true
	}

	nearestDistance := math.Inf(1)
	var nearestNormal vector3.Float64
	hitSomething := false

	grazedRim := false
	s.eachAlong(geometry.NewRay(origin, direction), 0, reach, func(faceIndex int) bool {
		distance, margin, hit := s.faces[faceIndex].rayHit(origin, direction)
		if !hit || distance <= s.tolerance || distance > reach {
			return true
		}
		// Landing on a rim means the two triangles sharing it disagree about
		// which way the surface faces here, so this ray settles nothing.
		if margin <= rim {
			grazedRim = true
			return false
		}
		if distance < nearestDistance {
			nearestDistance, nearestNormal, hitSomething = distance, s.faces[faceIndex].normal, true
		}
		return true
	})

	if grazedRim {
		return outside, false
	}
	if !hitSomething {
		return outside, true
	}
	if nearestNormal.Dot(direction) > 0 {
		return inside, true
	}
	return outside, true
}

// Small turns off the face normal, tried in order. The first is the paper's
// own direction; the rest only matter when that one grazes an edge.
func nudged(normal vector3.Float64) []vector3.Float64 {
	sideways := normal.Perpendicular().Normalized()
	upward := normal.Cross(sideways)

	directions := []vector3.Float64{normal}
	for _, turn := range []float64{0.013, -0.021, 0.037, -0.053} {
		directions = append(directions,
			normal.Add(sideways.Scale(turn)).Add(upward.Scale(turn*0.6)).Normalized())
	}
	return directions
}
