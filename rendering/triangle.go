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
	intersects := rayIntersectsTri(intersectingTri{
		p1: tri.p1,
		p2: tri.p2,
		p3: tri.p3,

		n1: tri.n1,
		n2: tri.n2,
		n3: tri.n3,
	}, ray.Ray(), minDistance, maxDistance, hitRecord)
	if !intersects {
		return false
	}

	hitRecord.Material = tri.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	// hitRecord.UV = s.UV(hitRecord.Normal)

	return true
}
