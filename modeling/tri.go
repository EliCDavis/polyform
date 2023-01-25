package modeling

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type scopedTri struct {
	data  []vector3.Float64
	p1    int
	p2    int
	p3    int
	plane *geometry.Plane
}

func (t scopedTri) Plane() geometry.Plane {
	if t.plane == nil {
		plane := geometry.NewPlaneFromPoints(
			t.data[t.p1],
			t.data[t.p2],
			t.data[t.p3],
		)
		t.plane = &plane
	}
	return *t.plane
}

// https://gdbooks.gitbooks.io/3dcollisions/content/Chapter4/point_in_triangle.html
func (t scopedTri) PointInSide(p vector3.Float64) bool {
	// Move the triangle so that the point becomes the
	// triangles origin
	a := t.data[t.p1].Sub(p)
	b := t.data[t.p2].Sub(p)
	c := t.data[t.p3].Sub(p)

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

func (t scopedTri) ClosestPoint(p vector3.Float64) vector3.Float64 {
	closestPoint := t.Plane().ClosestPoint(p)

	if t.PointInSide(closestPoint) {
		return closestPoint
	}

	AB := geometry.NewLine3D(t.data[t.p1], t.data[t.p2])
	BC := geometry.NewLine3D(t.data[t.p2], t.data[t.p3])
	CA := geometry.NewLine3D(t.data[t.p3], t.data[t.p1])

	c1 := AB.ClosestPointOnLine(closestPoint)
	c2 := BC.ClosestPointOnLine(closestPoint)
	c3 := CA.ClosestPointOnLine(closestPoint)

	mag1 := closestPoint.Sub(c1).LengthSquared()
	mag2 := closestPoint.Sub(c2).LengthSquared()
	mag3 := closestPoint.Sub(c3).LengthSquared()

	min := math.Min(mag1, mag2)
	min = math.Min(min, mag3)

	if min == mag1 {
		return c1
	} else if min == mag2 {
		return c2
	}
	return c3
}

func (t scopedTri) BoundingBox() AABB {
	aabb := NewAABB(t.data[t.p1], vector3.Zero[float64]())
	aabb.EncapsulatePoint(t.data[t.p2])
	aabb.EncapsulatePoint(t.data[t.p3])
	return aabb
}

// Tri provides utility functions to a specific underlying mesh
type Tri struct {
	mesh          *Mesh
	startingIndex int
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

func (t Tri) P1Vec3Attr(atr string) vector3.Float64 {
	return t.mesh.v3Data[atr][t.P1()]
}

func (t Tri) P2Vec3Attr(atr string) vector3.Float64 {
	return t.mesh.v3Data[atr][t.P2()]
}

func (t Tri) P3Vec3Attr(atr string) vector3.Float64 {
	return t.mesh.v3Data[atr][t.P3()]
}

func (t Tri) L1(atr string) geometry.Line3D {
	return geometry.NewLine3D(
		t.P1Vec3Attr(atr),
		t.P2Vec3Attr(atr),
	)
}

func (t Tri) L2(atr string) geometry.Line3D {
	return geometry.NewLine3D(
		t.P2Vec3Attr(atr),
		t.P3Vec3Attr(atr),
	)
}

func (t Tri) L3(atr string) geometry.Line3D {
	return geometry.NewLine3D(
		t.P3Vec3Attr(atr),
		t.P1Vec3Attr(atr),
	)
}

func (t Tri) Plane(attr string) geometry.Plane {
	return geometry.NewPlaneFromPoints(
		t.P1Vec3Attr(PositionAttribute),
		t.P2Vec3Attr(PositionAttribute),
		t.P3Vec3Attr(PositionAttribute),
	)
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

	aabb := NewAABB(center, vector3.Zero[float64]())
	aabb.EncapsulatePoint(t.P1Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P2Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P3Vec3Attr(PositionAttribute))

	return aabb
}

// https://gdbooks.gitbooks.io/3dcollisions/content/Chapter4/point_in_triangle.html
func (t Tri) PointInSide(p vector3.Float64) bool {
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

func (t Tri) ClosestPoint(atr string, p vector3.Float64) vector3.Float64 {
	closestPoint := t.Plane(atr).ClosestPoint(p)

	if t.PointInSide(closestPoint) {
		return closestPoint
	}

	AB := geometry.NewLine3D(t.P1Vec3Attr(atr), t.P2Vec3Attr(atr))
	BC := geometry.NewLine3D(t.P2Vec3Attr(atr), t.P3Vec3Attr(atr))
	CA := geometry.NewLine3D(t.P3Vec3Attr(atr), t.P1Vec3Attr(atr))

	c1 := AB.ClosestPointOnLine(closestPoint)
	c2 := BC.ClosestPointOnLine(closestPoint)
	c3 := CA.ClosestPointOnLine(closestPoint)

	mag1 := closestPoint.Sub(c1).LengthSquared()
	mag2 := closestPoint.Sub(c2).LengthSquared()
	mag3 := closestPoint.Sub(c3).LengthSquared()

	min := math.Min(mag1, mag2)
	min = math.Min(min, mag3)

	if min == mag1 {
		return c1
	} else if min == mag2 {
		return c2
	}
	return c3
}

func (t Tri) BoundingBox(atr string) AABB {
	aabb := NewAABB(t.P1Vec3Attr(atr), vector3.Zero[float64]())
	aabb.EncapsulatePoint(t.P2Vec3Attr(atr))
	aabb.EncapsulatePoint(t.P3Vec3Attr(atr))
	return aabb
}

func (t Tri) Scope(atr string) ScopedPrimitive {
	return &scopedTri{
		data: t.mesh.v3Data[atr],
		p1:   t.mesh.indices[t.startingIndex],
		p2:   t.mesh.indices[t.startingIndex+1],
		p3:   t.mesh.indices[t.startingIndex+2],
	}
}
