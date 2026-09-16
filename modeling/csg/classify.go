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

// Face tests a query budget may spend scanning before a tree is cheaper.
// Scanning a face costs far less than filing it, but the gap closes once the
// mesh outgrows the cache, so a fixed query count is wrong at both ends.
const scanBudget = 2_000_000

func newSolid(faces []face, tolerance float64, queries int) *solid {
	build := queries*len(faces) >= scanBudget

	var elements []trees.Element
	if build {
		elements = make([]trees.Element, len(faces))
	}

	// Grown from the face boxes already being built for the tree, rather
	// than from a second list holding every vertex in the mesh.
	var bounds geometry.AABB
	for i, f := range faces {
		box := f.bounds()
		if build {
			elements[i] = faceElement{bounds: box}
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
	if build {
		s.tree = trees.NewOctree(elements)
	}
	return s
}

// The tree only ever narrows the field; callers re-test whatever it hands
// back, so scanning everything reaches the same answer by a slower road.
// Visiting stops early when visit returns false.
func (s *solid) eachNear(origin vector3.Float64, radius float64, visit func(int) bool) {
	if s.tree == nil {
		s.eachFace(visit)
		return
	}
	for _, i := range s.tree.ElementsWithinRange(origin, radius) {
		if !visit(i) {
			return
		}
	}
}

func (s *solid) eachAlong(ray geometry.Ray, min, max float64, visit func(int) bool) {
	if s.tree == nil {
		s.eachFace(visit)
		return
	}
	for _, i := range s.tree.ElementsIntersectingRay(ray, min, max) {
		if !visit(i) {
			return
		}
	}
}

func (s *solid) eachFace(visit func(int) bool) {
	for i := range s.faces {
		if !visit(i) {
			return
		}
	}
}

// How far outside a triangle a hit may land and still count. A ray leaving
// through an edge two triangles share sits a rounding error to one side of
// both of them, and rejecting it from each in turn loses the crossing
// altogether: the ray then appears to leave the solid without touching it.
const rim = 1e-9

// Möller-Trumbore. edge reports how close the hit landed to the triangle's
// rim, which is what says whether the answer can be trusted.
func (f face) rayHit(origin, direction vector3.Float64) (distance, edge float64, ok bool) {
	e1 := f.verts[1].Sub(f.verts[0])
	e2 := f.verts[2].Sub(f.verts[0])

	pvec := direction.Cross(e2)
	det := e1.Dot(pvec)
	if det == 0 {
		return 0, 0, false
	}

	inv := 1 / det
	tvec := origin.Sub(f.verts[0])
	u := tvec.Dot(pvec) * inv
	qvec := tvec.Cross(e1)
	v := direction.Dot(qvec) * inv

	if u < -rim || v < -rim || u+v > 1+rim {
		return 0, 0, false
	}

	return e2.Dot(qvec) * inv, math.Min(u, math.Min(v, 1-u-v)), true
}

func (f face) planeDistance(p vector3.Float64) float64 {
	return p.Sub(f.verts[0]).Dot(f.normal)
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

	if boundary, ok := s.onBoundary(f, origin); ok {
		return boundary
	}

	for _, direction := range nudged(f.normal) {
		if result, ok := s.castFrom(origin, direction); ok {
			return result
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
func (s *solid) onBoundary(f face, origin vector3.Float64) (classification, bool) {
	answer, found := outside, false
	s.eachNear(origin, s.tolerance, func(i int) bool {
		other := s.faces[i]
		if math.Abs(other.planeDistance(origin)) > s.tolerance {
			return true
		}
		if _, _, hit := other.rayHit(origin, other.normal); !hit {
			return true
		}
		if f.normal.Dot(other.normal) > 0 {
			answer, found = same, true
		} else {
			answer, found = opposite, true
		}
		return false
	})
	return answer, found
}

// "If the normal to polygonB points toward polygonA, then polygonA is OUTSIDE
// objectB; otherwise, polygonA is INSIDE objectB. If no polygons were
// intersected, then polygonA is OUTSIDE objectB."
func (s *solid) castFrom(origin, direction vector3.Float64) (classification, bool) {
	reach := s.bounds.Size().Length() * 2
	if reach == 0 {
		return outside, true
	}

	nearest := math.Inf(1)
	var nearestNormal vector3.Float64
	found := false

	grazed := false
	s.eachAlong(geometry.NewRay(origin, direction), 0, reach, func(i int) bool {
		distance, edge, ok := s.faces[i].rayHit(origin, direction)
		if !ok || distance <= s.tolerance || distance > reach {
			return true
		}
		// Landing on a rim means the two triangles sharing it disagree about
		// which way the surface faces here, so this ray settles nothing.
		if edge <= rim {
			grazed = true
			return false
		}
		if distance < nearest {
			nearest, nearestNormal, found = distance, s.faces[i].normal, true
		}
		return true
	})

	if grazed {
		return outside, false
	}
	if !found {
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
	other := normal.Cross(sideways)

	out := []vector3.Float64{normal}
	for _, turn := range []float64{0.013, -0.021, 0.037, -0.053} {
		out = append(out,
			normal.Add(sideways.Scale(turn)).Add(other.Scale(turn*0.6)).Normalized())
	}
	return out
}
