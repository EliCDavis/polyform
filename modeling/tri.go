package modeling

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
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

func (t scopedTri) BoundingBox() geometry.AABB {
	return geometry.NewAABBFromPoints(
		t.data[t.p1],
		t.data[t.p2],
		t.data[t.p3],
	)
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

func (t Tri) P1Vec3Attr(attr string) vector3.Float64 {
	return t.mesh.v3Data[attr][t.P1()]
}

func (t Tri) P2Vec3Attr(attr string) vector3.Float64 {
	return t.mesh.v3Data[attr][t.P2()]
}

func (t Tri) P3Vec3Attr(attr string) vector3.Float64 {
	return t.mesh.v3Data[attr][t.P3()]
}

func (t Tri) P1Vec2Attr(attr string) vector2.Float64 {
	return t.mesh.v2Data[attr][t.P1()]
}

func (t Tri) P2Vec2Attr(attr string) vector2.Float64 {
	return t.mesh.v2Data[attr][t.P2()]
}

func (t Tri) P3Vec2Attr(attr string) vector2.Float64 {
	return t.mesh.v2Data[attr][t.P3()]
}

func (t Tri) P1Vec1Attr(attr string) float64 {
	return t.mesh.v1Data[attr][t.P1()]
}

func (t Tri) P2Vec1Attr(attr string) float64 {
	return t.mesh.v1Data[attr][t.P2()]
}

func (t Tri) P3Vec1Attr(attr string) float64 {
	return t.mesh.v1Data[attr][t.P3()]
}

func (t Tri) L1(attr string) geometry.Line3D {
	return geometry.NewLine3D(
		t.P1Vec3Attr(attr),
		t.P2Vec3Attr(attr),
	)
}

func (t Tri) L2(attr string) geometry.Line3D {
	return geometry.NewLine3D(
		t.P2Vec3Attr(attr),
		t.P3Vec3Attr(attr),
	)
}

func (t Tri) L3(attr string) geometry.Line3D {
	return geometry.NewLine3D(
		t.P3Vec3Attr(attr),
		t.P1Vec3Attr(attr),
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

func (t Tri) Bounds() geometry.AABB {
	center := t.P1Vec3Attr(PositionAttribute).
		Add(t.P2Vec3Attr(PositionAttribute)).
		Add(t.P3Vec3Attr(PositionAttribute)).
		DivByConstant(3)

	aabb := geometry.NewAABB(center, vector3.Zero[float64]())
	aabb.EncapsulatePoint(t.P1Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P2Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P3Vec3Attr(PositionAttribute))

	return aabb
}

func (t Tri) Average(attr string) vector3.Float64 {
	return t.P1Vec3Attr(attr).
		Add(t.P2Vec3Attr(attr)).
		Add(t.P3Vec3Attr(attr)).
		Scale(1. / 3.)
}

// https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/moller-trumbore-ray-triangle-intersection.html
func (t Tri) RayIntersects(ray geometry.Ray) (vector3.Float64, bool) {
	const kEpsilon = 0.00001

	dir := ray.Direction()
	orig := ray.Origin()
	v0 := t.P1Vec3Attr(PositionAttribute)
	v1 := t.P2Vec3Attr(PositionAttribute)
	v2 := t.P3Vec3Attr(PositionAttribute)

	v0v1 := v1.Sub(v0)
	v0v2 := v2.Sub(v0)
	pvec := dir.Cross(v0v2)
	det := v0v1.Dot(pvec)

	// ray and triangle are parallel if det is close to 0
	if math.Abs(det) < kEpsilon {
		return vector3.Zero[float64](), false
	}

	invDet := 1. / det

	tvec := orig.Sub(v0)
	u := tvec.Dot(pvec) * invDet
	if u < 0 || u > 1 {
		return vector3.Zero[float64](), false
	}

	qvec := tvec.Cross(v0v1)
	v := dir.Dot(qvec) * invDet
	if v < 0 || u+v > 1 {
		return vector3.Zero[float64](), false
	}

	tVal := v0v2.Dot(qvec) * invDet

	return ray.At(tVal), true
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

func (t Tri) LineIntersects(line geometry.Line3D) (vector3.Float64, bool) {
	plane := t.Plane(PositionAttribute)
	point, intersects := line.IntersectionPointOnPlane(plane)
	if !intersects {
		return vector3.Zero[float64](), false
	}
	if t.PointInSide(point) {
		return point, true
	}
	return vector3.Zero[float64](), false
}

func (t Tri) ClosestPoint(attr string, p vector3.Float64) vector3.Float64 {
	closestPoint := t.Plane(attr).ClosestPoint(p)

	if t.PointInSide(closestPoint) {
		return closestPoint
	}

	AB := geometry.NewLine3D(t.P1Vec3Attr(attr), t.P2Vec3Attr(attr))
	BC := geometry.NewLine3D(t.P2Vec3Attr(attr), t.P3Vec3Attr(attr))
	CA := geometry.NewLine3D(t.P3Vec3Attr(attr), t.P1Vec3Attr(attr))

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

func (t Tri) BoundingBox(attr string) geometry.AABB {
	aabb := geometry.NewAABB(t.P1Vec3Attr(attr), vector3.Zero[float64]())
	aabb.EncapsulatePoint(t.P2Vec3Attr(attr))
	aabb.EncapsulatePoint(t.P3Vec3Attr(attr))
	return aabb
}

func (t Tri) Area3D(attr string) float64 {
	p1 := t.P1Vec3Attr(attr)
	p2 := t.P2Vec3Attr(attr)
	p3 := t.P3Vec3Attr(attr)

	return p2.Sub(p1).Cross(p3.Sub(p1)).Length() / 2
}

func (t Tri) Scope(attr string) trees.Element {
	return &scopedTri{
		data: t.mesh.v3Data[attr],
		p1:   t.mesh.indices[t.startingIndex],
		p2:   t.mesh.indices[t.startingIndex+1],
		p3:   t.mesh.indices[t.startingIndex+2],
	}
}
