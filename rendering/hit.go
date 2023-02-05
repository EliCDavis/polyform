package rendering

import (
	"github.com/EliCDavis/vector/vector3"
)

type HitList []Hittable

func (h HitList) Hit(r *TemporalRay, min, max float64, hitRecord *HitRecord) bool {
	tempRecord := NewHitRecord()
	hitAnything := false
	closestSoFar := max

	for i := 0; i < len(h); i++ {
		if h[i].Hit(r, min, closestSoFar, tempRecord) {
			hitAnything = true
			closestSoFar = tempRecord.Distance

			*hitRecord = *tempRecord
		}
	}

	return hitAnything
}

type Hittable interface {
	Hit(r *TemporalRay, min, max float64, hitRecord *HitRecord) bool
}

type HitRecord struct {
	Distance  float64
	Point     vector3.Float64
	Normal    vector3.Float64
	FrontFace bool
	Material  Material
}

func NewHitRecord() *HitRecord {
	return &HitRecord{
		0,
		vector3.Zero[float64](),
		vector3.Zero[float64](),
		true,
		nil,
	}
}

func (h *HitRecord) SetFaceNormal(ray TemporalRay, outwardNormal vector3.Float64) {
	h.FrontFace = ray.Direction().Dot(outwardNormal) < 0
	if h.FrontFace {
		h.Normal = outwardNormal
	} else {
		h.Normal = outwardNormal.Scale(-1.)
	}
}
