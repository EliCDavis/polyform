package rendering

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type Triangle struct {
	p1, p2, p3 vector3.Float64
	n1, n2, n3 vector3.Float64
	mat        Material
	box        geometry.AABB
}

func (tri Triangle) BoundingBox(start, stop float64) *geometry.AABB {
	return &tri.box
}

func (tri Triangle) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	point, intersects := rayIntersectsTri(tri.p1, tri.p2, tri.p3, *ray)
	if !intersects {
		return false
	}

	distFromOrig := point.Distance(ray.Origin())

	// Prevents us from bouncing around and dying inside the triangle itself
	if distFromOrig < 0.001 {
		return false
	}

	if distFromOrig > maxDistance {
		return false
	}

	// compute the plane's normal
	v0v1 := tri.p2.Sub(tri.p1)
	v0v2 := tri.p3.Sub(tri.p1)
	// no need to normalize
	N := v0v1.Cross(v0v2) // N
	denom := N.Dot(N)

	edge1 := tri.p3.Sub(tri.p2)
	vp1 := point.Sub(tri.p2)
	C := edge1.Cross(vp1)
	u := N.Dot(C)

	edge2 := tri.p1.Sub(tri.p3)
	vp2 := point.Sub(tri.p3)
	C = edge2.Cross(vp2)
	v := N.Dot(C)

	u /= denom
	v /= denom

	w := 1. - u - v
	normal := tri.n1.Scale(u).
		Add(tri.n2.Scale(v)).
		Add(tri.n3.Scale(w)).
		Normalized()

	hitRecord.Normal = normal
	hitRecord.Distance = distFromOrig
	hitRecord.Point = point
	hitRecord.Float3Data["barycentric"] = vector3.New(u, v, w)
	hitRecord.Material = tri.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	// hitRecord.UV = s.UV(hitRecord.Normal)

	return true
}
