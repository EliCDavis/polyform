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
	corners := geometry.Triangle{tri.p1, tri.p2, tri.p3}
	hit, ok := corners.RayHit(ray.Ray())
	if !accepted(hit, ok, minDistance, maxDistance) {
		return false
	}
	recordHit(hitRecord, hit, corners, ray.Ray())

	hitRecord.Material = tri.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	// hitRecord.UV = s.UV(hitRecord.Normal)

	return true
}
