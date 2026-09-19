package modeling

import (
	"math/rand/v2"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

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

func LowDiscrepancySample(u float32) (float64, float64) {
	// https://pharr.org/matt/blog/2019/03/13/triangle-sampling-1.5.html

	uf := uint32(u * (1 << 32))
	cx, cy := float32(0.0), float32(0.0)
	w := float32(0.5)

	for range 16 {
		uu := uf >> 30
		flip := (uu & 3) == 0

		if (uu & 1) == 0 {
			cy += w
		}
		if (uu & 2) == 0 {
			cx += w
		}

		if flip {
			w *= -0.5
		} else {
			w *= 0.5
		}
		uf <<= 2
	}

	return float64(cx + w/3.0), float64(cy + w/3.0)
}

func (t Tri) UniformSample(attr string, in vector2.Float64) vector3.Float64 {
	// https://pharr.org/matt/blog/2019/03/13/triangle-sampling-1.5
	u := in.X() / 2.
	v := in.Y() / 2.
	offset := v - u
	if offset > 0 {
		v += offset
	} else {
		u -= offset
	}
	return t.P1Vec3Attr(attr).Scale(1. - u - v).
		Add(t.P2Vec3Attr(attr).Scale(u)).
		Add(t.P3Vec3Attr(attr).Scale(v))
}

func (t Tri) RandomSample(attr string, rand *rand.Rand) vector3.Float64 {
	// TODO: Implement like, something actually smart
	// See resources in README for quasirandom for actual smart stuff

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Dumbass Attempt #1
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// a := t.P1Vec3Attr(attr)
	// b := t.P2Vec3Attr(attr)
	// c := t.P3Vec3Attr(attr)

	// x := vector3.Lerp(a, b, rand.Float64())
	// return vector3.Lerp(x, c, rand.Float64())
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Attempt #2
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// a := t.P1Vec3Attr(attr)
	// b := t.P2Vec3Attr(attr)
	// c := t.P3Vec3Attr(attr)
	// ab := b.Sub(a)
	// ac := c.Sub(a)

	// r1 := rand.Float64()
	// r2 := rand.Float64()

	// // Sample a square
	// p := ab.Scale(r1).Add(ac.Scale(r2)).Add(a)

	// // Reflect if it's not inside our triangle
	// if !t.PointInSide(p) {
	// 	p = ab.Scale(1 - r1).Add(ac.Scale(1 - r2)).Add(a)
	// }

	// return p
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	u, v := LowDiscrepancySample(float32(rand.Float64()))
	return t.P1Vec3Attr(attr).Scale(1. - u - v).
		Add(t.P2Vec3Attr(attr).Scale(u)).
		Add(t.P3Vec3Attr(attr).Scale(v))
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

// Triangle is the corners' values of a float3 attribute, as a geometry.
func (t Tri) Triangle(attr string) geometry.Triangle {
	return geometry.Triangle{t.P1Vec3Attr(attr), t.P2Vec3Attr(attr), t.P3Vec3Attr(attr)}
}

func (t Tri) RayIntersects(ray geometry.Ray) (vector3.Float64, bool) {
	tri := t.Triangle(PositionAttribute)
	hit, ok := tri.RayHit(ray)
	if !ok || !hit.Inside(0) {
		return vector3.Zero[float64](), false
	}
	return ray.At(hit.Distance), true
}

func (t Tri) Normal(attr string) vector3.Float64 {
	return t.Triangle(attr).Normal()
}

// PointInSide reports whether p, dropped onto the triangle's plane, lands
// within its edges.
func (t Tri) PointInSide(p vector3.Float64) bool {
	return t.Triangle(PositionAttribute).ProjectsInside(p)
}

func (t Tri) LineIntersects(line geometry.Line3D) (vector3.Float64, bool) {
	point, intersects := line.IntersectionPointOnPlane(t.Plane(PositionAttribute))
	if !intersects || !t.PointInSide(point) {
		return vector3.Zero[float64](), false
	}
	return point, true
}

func (t Tri) ClosestPoint(attr string, p vector3.Float64) vector3.Float64 {
	return t.Triangle(attr).ClosestPoint(p)
}

func (t Tri) BoundingBox(attr string) geometry.AABB {
	return t.Triangle(attr).BoundingBox()
}

func (t Tri) Area3D(attr string) float64 {
	return t.Triangle(attr).Area()
}

func (t Tri) Scope(attr string) trees.Element {
	return t.Triangle(attr)
}
