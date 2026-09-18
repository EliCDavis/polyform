package geometry

import (
	"math"

	"github.com/EliCDavis/polyform/math/predicate"
	"github.com/EliCDavis/vector/vector3"
)

type Triangle [3]vector3.Float64

// Normal is unit length, or zero for a triangle with no area.
func (t Triangle) Normal() vector3.Float64 {
	normal := t[1].Sub(t[0]).Cross(t[2].Sub(t[0]))
	length := normal.Length()
	if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
		return vector3.Zero[float64]()
	}
	return normal.DivByConstant(length)
}

func (t Triangle) Area() float64 {
	return t[1].Sub(t[0]).Cross(t[2].Sub(t[0])).Length() / 2
}

func (t Triangle) Translate(amount vector3.Float64) Triangle {
	return Triangle{t[0].Add(amount), t[1].Add(amount), t[2].Add(amount)}
}

func (t Triangle) Centroid() vector3.Float64 {
	return t[0].Add(t[1]).Add(t[2]).Scale(1. / 3.)
}

func (t Triangle) BoundingBox() AABB {
	return NewAABBFromPoints(t[0], t[1], t[2])
}

func (t Triangle) Plane() Plane {
	return NewPlane(t[0], t.Normal())
}

// Reach is the farthest any corner sits from the centroid. Half the bounding
// box diagonal can be smaller and would miss points on the triangle's own edges.
func (t Triangle) Reach() float64 {
	center := t.Centroid()
	return max(
		t[0].Sub(center).Length(),
		t[1].Sub(center).Length(),
		t[2].Sub(center).Length(),
	)
}

// Degenerate reports a triangle whose height is within tolerance of zero.
func (t Triangle) Degenerate(tolerance float64) bool {
	longest := max(t[0].Distance(t[1]), t[1].Distance(t[2]), t[2].Distance(t[0]))
	return 2*t.Area() <= tolerance*longest
}

// Contains reports whether point lies on the triangle, within tolerance of
// its plane and its edges. A triangle with no area contains nothing.
func (t Triangle) Contains(point vector3.Float64, tolerance float64) bool {
	normal := t.Normal()
	if normal.Length() == 0 || math.Abs(point.Sub(t[0]).Dot(normal)) > tolerance {
		return false
	}
	for i := 0; i < 3; i++ {
		edge := t[(i+1)%3].Sub(t[i])
		if edge.Cross(point.Sub(t[i])).Dot(normal) < -tolerance*edge.Length() {
			return false
		}
	}
	return true
}

// ClosestPointOnEdges is the nearest point on the triangle's outline, and
// how far away it is.
func (t Triangle) ClosestPointOnEdges(point vector3.Float64) (vector3.Float64, float64) {
	nearest, nearestDistance := t[0], math.Inf(1)
	for i := 0; i < 3; i++ {
		onEdge := NewLine3D(t[i], t[(i+1)%3]).ClosestPointOnLine(point)
		if distance := onEdge.Distance(point); distance < nearestDistance {
			nearest, nearestDistance = onEdge, distance
		}
	}
	return nearest, nearestDistance
}

// ProjectsInside reports whether point, dropped straight onto the plane,
// lands within the edges. A triangle with no area has no inside.
func (t Triangle) ProjectsInside(point vector3.Float64) bool {
	normal := t.Normal()
	if normal.Length() == 0 {
		return false
	}
	for i := 0; i < 3; i++ {
		edge := t[(i+1)%3].Sub(t[i])
		if edge.Cross(point.Sub(t[i])).Dot(normal) < 0 {
			return false
		}
	}
	return true
}

// ClosestPoint is the nearest point on the triangle's surface.
func (t Triangle) ClosestPoint(point vector3.Float64) vector3.Float64 {
	if !t.ProjectsInside(point) {
		nearest, _ := t.ClosestPointOnEdges(point)
		return nearest
	}
	normal := t.Normal()
	return point.Sub(normal.Scale(point.Sub(t[0]).Dot(normal)))
}

// SidesOf is which side of other's plane each corner falls on, as Orient3D
// gives it: exact in sign, scaled by twice other's area.
func (t Triangle) SidesOf(other Triangle) [3]float64 {
	var sides [3]float64
	for i, corner := range t {
		sides[i] = predicate.Orient3D(other[0], other[1], other[2], corner)
	}
	return sides
}

// Coplanar reports whether every corner lies within tolerance of other's
// plane.
func (t Triangle) Coplanar(other Triangle, tolerance float64) bool {
	return other.Plane().Holds(t, tolerance)
}

// Intersect is the segment two triangles share when they cross. Coplanar
// pairs share none, and an overlap shorter than tolerance does not count.
func (t Triangle) Intersect(other Triangle, tolerance float64) (Line3D, bool) {
	sides := t.SidesOf(other)
	otherSides := other.SidesOf(t)

	if entirelyOneSide(sides) || entirelyOneSide(otherSides) {
		return Line3D{}, false
	}
	// Any closer and the direction below is a cross product of rounding error.
	if t.Coplanar(other, tolerance) || other.Coplanar(t, tolerance) {
		return Line3D{}, false
	}

	crossing, ok := t.crossesPlane(sides, tolerance)
	if !ok {
		return Line3D{}, false
	}
	otherCrossing, ok := other.crossesPlane(otherSides, tolerance)
	if !ok {
		return Line3D{}, false
	}

	// Both segments lie on the line where the two planes meet, so they can be
	// compared as intervals along it.
	direction := t.Normal().Cross(other.Normal())
	if direction.Length() == 0 {
		return Line3D{}, false
	}
	direction = direction.Normalized()

	base := crossing[0]
	along := func(point vector3.Float64) float64 { return point.Sub(base).Dot(direction) }

	// The overlap ends on two of the four crossing points already computed.
	// Rebuilding one from a parameter would round it differently.
	ordered := func(pair [2]vector3.Float64) (start, end vector3.Float64, startAt, endAt float64) {
		first, second := along(pair[0]), along(pair[1])
		if first <= second {
			return pair[0], pair[1], first, second
		}
		return pair[1], pair[0], second, first
	}

	start, end, startAt, endAt := ordered(crossing)
	otherStart, otherEnd, otherStartAt, otherEndAt := ordered(otherCrossing)

	if otherStartAt > startAt {
		start, startAt = otherStart, otherStartAt
	}
	if otherEndAt < endAt {
		end, endAt = otherEnd, otherEndAt
	}

	if endAt-startAt <= tolerance {
		return Line3D{}, false
	}
	return NewLine3D(start, end), true
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

// Where the triangle meets a plane: two points, or nothing when it only
// touches at a corner. sides is which side of the plane each corner is on.
func (t Triangle) crossesPlane(sides [3]float64, tolerance float64) ([2]vector3.Float64, bool) {
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
			record(t[i])
			continue
		}
		if endSide == 0 || (startSide > 0) == (endSide > 0) {
			continue
		}

		// The area factor in both cancels, leaving the true fraction.
		fraction := startSide / (startSide - endSide)
		record(t[i].Add(t[j].Sub(t[i]).Scale(fraction)))
	}

	if len(crossings) != 2 {
		return [2]vector3.Float64{}, false
	}
	return [2]vector3.Float64{crossings[0], crossings[1]}, true
}
