package modeling

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector"
)

// Tri provides utility functions to a specific underlying mesh
type Tri struct {
	mesh          *Mesh
	startingIndex int
	plane         *Plane
}

// P1 is the first point on our triangle, which is an index to the vertices array of a mesh
func (t Tri) P1() int {
	return t.mesh.indices[t.startingIndex]
}

// P2 is the second point on our triangle, which is an index to the vertices array of a mesh
func (t Tri) P2() int {
	return t.mesh.indices[t.startingIndex+1]
}

// P3 is the third point on our triangle, which is an index to the vertices array of a mesh
func (t Tri) P3() int {
	return t.mesh.indices[t.startingIndex+2]
}

func (t Tri) P1Vec3Attr(attr string) vector.Vector3 {
	return t.mesh.v3Data[attr][t.P1()]
}

func (t Tri) P2Vec3Attr(attr string) vector.Vector3 {
	return t.mesh.v3Data[attr][t.P2()]
}

func (t Tri) P3Vec3Attr(attr string) vector.Vector3 {
	return t.mesh.v3Data[attr][t.P3()]
}

func (t Tri) Plane() Plane {
	if t.plane == nil {
		plane := NewPlaneFromPoints(
			t.P1Vec3Attr(PositionAttribute),
			t.P2Vec3Attr(PositionAttribute),
			t.P3Vec3Attr(PositionAttribute),
		)
		t.plane = &plane
	}
	return *t.plane
}

// Valid determines whether or not the contains 3 unique vertices.
func (t Tri) UniqueVertices() bool {
	if t.P1() == t.P2() {
		return false
	}
	if t.P1() == t.P3() {
		return false
	}
	if t.P2() == t.P3() {
		return false
	}
	return true
}

func (t Tri) Bounds() AABB {

	center := t.P1Vec3Attr(PositionAttribute).
		Add(t.P2Vec3Attr(PositionAttribute)).
		Add(t.P3Vec3Attr(PositionAttribute)).
		DivByConstant(3)

	aabb := NewAABB(center, vector.Vector3Zero())
	aabb.EncapsulatePoint(t.P1Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P2Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P3Vec3Attr(PositionAttribute))

	return aabb
}

// https://gdbooks.gitbooks.io/3dcollisions/content/Chapter4/point_in_triangle.html
func (t Tri) PointInSide(p vector.Vector3) bool {
	// Move the triangle so that the point becomes the
	// triangles origin
	a := t.P1Vec3Attr(PositionAttribute).Sub(p)
	b := t.P2Vec3Attr(PositionAttribute).Sub(p)
	c := t.P3Vec3Attr(PositionAttribute).Sub(p)

	// Compute the normal vectors for triangles:
	// u = normal of PBC
	// v = normal of PCA
	// w = normal of PAB

	u := b.Cross(c)
	v := c.Cross(a)

	// Test to see if the normals are facing
	// the same direction, return false if not
	if u.Dot(v) < 0. {
		return false
	}

	w := a.Cross(b)
	return u.Dot(w) >= 0.
}

func (t Tri) ClosestPoint(p vector.Vector3) vector.Vector3 {

	closestPoint := t.Plane().ClosestPoint(p)

	if t.PointInSide(closestPoint) {
		return closestPoint
	}

	AB := geometry.NewLine3D(t.P1Vec3Attr(PositionAttribute), t.P2Vec3Attr(PositionAttribute))
	BC := geometry.NewLine3D(t.P2Vec3Attr(PositionAttribute), t.P3Vec3Attr(PositionAttribute))
	CA := geometry.NewLine3D(t.P3Vec3Attr(PositionAttribute), t.P1Vec3Attr(PositionAttribute))

	c1 := AB.ClosestPointOnLine(closestPoint)
	c2 := BC.ClosestPointOnLine(closestPoint)
	c3 := CA.ClosestPointOnLine(closestPoint)

	mag1 := closestPoint.Sub(c1).SquaredLength()
	mag2 := closestPoint.Sub(c2).SquaredLength()
	mag3 := closestPoint.Sub(c3).SquaredLength()

	min := math.Min(mag1, mag2)
	min = math.Min(min, mag3)

	if min == mag1 {
		return c1
	} else if min == mag2 {
		return c2
	}
	return c3
}
