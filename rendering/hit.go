package rendering

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
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

func (h HitList) BoundingBox(startTime, endTime float64) *geometry.AABB {
	if len(h) == 0 {
		return nil
	}

	box := geometry.NewAABB(vector3.Zero[float64](), vector3.Zero[float64]())

	hasBox := false

	for _, item := range h {
		itemBox := item.BoundingBox(startTime, endTime)
		if itemBox == nil {
			continue
		}
		box.EncapsulateBounds(*itemBox)
		hasBox = true
	}

	if !hasBox {
		return nil
	}

	return &box
}

type Hittable interface {
	Hit(r *TemporalRay, min, max float64, hitRecord *HitRecord) bool
	BoundingBox(startTime, endTime float64) *geometry.AABB
}

type HitRecord struct {
	Distance   float64
	Point      vector3.Float64
	Normal     vector3.Float64
	FrontFace  bool
	Material   Material
	UV         vector2.Float64
	Float3Data map[string]vector3.Float64
}

func NewHitRecord() *HitRecord {
	return &HitRecord{
		Distance:   0,
		Point:      vector3.Zero[float64](),
		Normal:     vector3.Zero[float64](),
		FrontFace:  true,
		Material:   nil,
		UV:         vector2.Zero[float64](),
		Float3Data: make(map[string]vector3.Float64),
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
