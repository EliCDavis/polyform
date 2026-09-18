package csg

import (
	"math"
	"sync"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling/winding"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

// The surface rays are cast against
type target struct {
	faces     []face
	tree      *trees.OctTree
	bounds    geometry.AABB
	tolerance float64

	// Only the fallback needs it, and it copies every face.
	windingOnce sync.Once
	winding     *winding.TriangleSampler
}

// Number of faces heuristic for when we should swap over to an octree.
const scanBudget = 2_000_000

func newTarget(faces []face, tolerance float64, queries int) *target {
	worthIndexing := queries*len(faces) >= scanBudget

	var boxes []trees.Element
	if worthIndexing {
		boxes = make([]trees.Element, len(faces))
	}

	var bounds geometry.AABB
	for i, f := range faces {
		box := f.verts.BoundingBox()
		if worthIndexing {
			boxes[i] = faceElement{bounds: box}
		}
		if i == 0 {
			bounds = box
			continue
		}
		bounds.EncapsulateBounds(box)
	}

	s := &target{
		faces:     faces,
		bounds:    bounds,
		tolerance: tolerance,
	}
	if worthIndexing {
		s.tree = trees.NewOctree(boxes)
	}
	return s
}

// The tree only narrows the field and callers re-test what it returns, so
// scanning every face gives the same answer. Stops when visit returns false.
func (s *target) eachNear(origin vector3.Float64, radius float64, visit func(faceIndex int) bool) {
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

func (s *target) eachAlong(ray geometry.Ray, min, max float64, visit func(faceIndex int) bool) {
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

func (s *target) eachFace(visit func(faceIndex int) bool) {
	for faceIndex := range s.faces {
		if !visit(faceIndex) {
			return
		}
	}
}

// How far outside a triangle a hit may land and still count. A ray leaving
// through a shared edge sits a rounding error outside both triangles.
const rim = 1e-9

// margin is how close the hit landed to the triangle's rim, in barycentric
// terms, which is what says whether it can be trusted.
func (f *face) rayHit(ray geometry.Ray) (distance, margin float64, hit bool) {
	at, ok := f.verts.RayHit(ray)
	if !ok || !at.Inside(rim) {
		return 0, 0, false
	}
	return at.Distance, at.Margin(), true
}

// Section 7: a ray from the face's barycenter along its normal, against the
// nearest face of the other solid. A hit too near a rim is retried nudged.
func (s *target) classify(f face) classification {
	origin := f.verts.Centroid()

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
// along.
func (s *target) enclosed(origin vector3.Float64) classification {
	s.windingOnce.Do(func() {
		tris := make([]geometry.Triangle, len(s.faces))
		for i, f := range s.faces {
			tris[i] = f.verts
		}
		s.winding = winding.FromTriangles(tris)
	})
	if s.winding.Inside(origin) {
		return inside
	}
	return outside
}

// Section 7: a face lying in the other surface is SAME or OPPOSITE by its
// normal. touchesSurface means only the barycenter is on it, not the face.
func (s *target) onBoundary(f face, origin vector3.Float64) (verdict classification, liesOnSurface, touchesSurface bool) {
	s.eachNear(origin, s.tolerance, func(faceIndex int) bool {
		other := &s.faces[faceIndex]
		plane := other.plane()
		if math.Abs(plane.Distance(origin)) > s.tolerance {
			return true
		}
		if _, _, hit := other.rayHit(geometry.NewRay(origin, other.normal)); !hit {
			return true
		}
		if !plane.Holds(f.verts, s.tolerance) {
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

// Section 7: outside when the nearest hit's normal faces the ray or nothing
// is hit, inside otherwise. settled is false when the ray grazed a rim.
func (s *target) castFrom(origin, direction vector3.Float64) (verdict classification, settled bool) {
	reach := s.bounds.Size().Length() * 2
	if reach == 0 {
		return outside, true
	}

	nearestDistance := math.Inf(1)
	var nearestNormal vector3.Float64
	hitSomething := false

	grazedRim := false
	ray := geometry.NewRay(origin, direction)
	s.eachAlong(ray, 0, reach, func(faceIndex int) bool {
		distance, margin, hit := s.faces[faceIndex].rayHit(ray)
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
