package geometry

import (
	"encoding/json"
	"math"

	"github.com/EliCDavis/vector/vector3"
)

type AABB struct {
	center  vector3.Float64
	extents vector3.Float64
}

func (aabb AABB) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Center  vector3.Float64 `json:"center"`
		Extents vector3.Float64 `json:"extents"`
	}{
		Center:  aabb.center,
		Extents: aabb.extents,
	})
}

func (aabb *AABB) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Center  vector3.Float64 `json:"center"`
		Extents vector3.Float64 `json:"extents"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	aabb.center = aux.Center
	aabb.extents = aux.Extents
	return nil
}

func NewAABB(center, size vector3.Float64) AABB {
	return AABB{
		center:  center,
		extents: size.Scale(0.5),
	}
}

func NewEmptyAABB() AABB {
	return AABB{
		center:  vector3.Zero[float64](),
		extents: vector3.Zero[float64](),
	}
}

func NewAABBFromPoints(points ...vector3.Float64) AABB {
	min := vector3.New(math.Inf(1), math.Inf(1), math.Inf(1))
	max := vector3.New(math.Inf(-1), math.Inf(-1), math.Inf(-1))
	for _, v := range points {
		min = min.SetX(math.Min(v.X(), min.X()))
		min = min.SetY(math.Min(v.Y(), min.Y()))
		min = min.SetZ(math.Min(v.Z(), min.Z()))

		max = max.SetX(math.Max(v.X(), max.X()))
		max = max.SetY(math.Max(v.Y(), max.Y()))
		max = max.SetZ(math.Max(v.Z(), max.Z()))
	}

	area := max.
		Sub(min)

	center := area.
		Scale(0.5).
		Add(min)

	return NewAABB(center, area)
}

func (aabb AABB) Center() vector3.Float64 {
	return aabb.center
}

func (aabb AABB) Min() vector3.Float64 {
	return aabb.center.Sub(aabb.extents)
}

func (aabb AABB) Max() vector3.Float64 {
	return aabb.center.Add(aabb.extents)
}

func (aabb AABB) Size() vector3.Float64 {
	return aabb.extents.Scale(2)
}

func (aabb AABB) Volume() float64 {
	size := aabb.Size()
	return size.X() * size.Y() * size.Z()
}

func (aabb AABB) Intersects(other AABB) bool {
	aMin := aabb.Min()
	aMax := aabb.Max()
	bMin := other.Min()
	bMax := other.Max()
	return (aMin.X() <= bMax.X()) && (aMax.X() >= bMin.X()) &&
		(aMin.Y() <= bMax.Y()) && (aMax.Y() >= bMin.Y()) &&
		(aMin.Z() <= bMax.Z()) && (aMax.Z() >= bMin.Z())
}

func (aabb *AABB) Expand(amount float64) {
	a := amount * 0.5
	aabb.extents = aabb.extents.Add(vector3.New(a, a, a))
}

func (aabb AABB) Contains(p vector3.Float64) bool {
	min := aabb.Min()
	if p.X() < min.X() {
		return false
	}
	if p.Y() < min.Y() {
		return false
	}
	if p.Z() < min.Z() {
		return false
	}

	max := aabb.Max()
	if p.X() > max.X() {
		return false
	}
	if p.Y() > max.Y() {
		return false
	}
	if p.Z() > max.Z() {
		return false
	}

	return true
}

func minVector(a, b vector3.Float64) vector3.Float64 {
	return vector3.New(
		math.Min(a.X(), b.X()),
		math.Min(a.Y(), b.Y()),
		math.Min(a.Z(), b.Z()),
	)
}

func maxVector(a, b vector3.Float64) vector3.Float64 {
	return vector3.New(
		math.Max(a.X(), b.X()),
		math.Max(a.Y(), b.Y()),
		math.Max(a.Z(), b.Z()),
	)
}

func (aabb *AABB) SetMinMax(min, max vector3.Float64) {
	aabb.extents = max.Sub(min).Scale(0.5)
	aabb.center = min.Add(aabb.extents)
}

func (aabb *AABB) EncapsulatePoint(p vector3.Float64) {
	aabb.SetMinMax(
		minVector(aabb.Min(), p),
		maxVector(aabb.Max(), p),
	)
}

func (aabb *AABB) EncapsulateBounds(b AABB) {
	aabb.EncapsulatePoint(b.center.Sub(b.extents))
	aabb.EncapsulatePoint(b.center.Add(b.extents))
}

func (aabb AABB) ClosestPoint(v vector3.Float64) vector3.Float64 {
	result := v
	min := aabb.Min()
	max := aabb.Max()
	result = result.SetX(clamp(v.X(), min.X(), max.X()))
	result = result.SetY(clamp(v.Y(), min.Y(), max.Y()))
	result = result.SetZ(clamp(v.Z(), min.Z(), max.Z()))
	return result
}

func clamp(v, min, max float64) float64 {
	return math.Min(math.Max(v, min), max)
}

// IntersectsRayInRange determines whether or not a ray intersects the bounding
// box with the range of the ray between min and max
//
// Intersection method by Andrew Kensler at Pixar, found in the book "Ray
// Tracing The Next Week" by Peter Shirley
func (aabb AABB) IntersectsRayInRange(ray Ray, min, max float64) bool {
	const kEpsilon = 0.0000000001
	boxMin := aabb.Min()
	boxMax := aabb.Max()

	rayMin := min
	rayMax := max
	if aabb.intersectsRayInRangeComponent(ray.origin.X(), ray.direction.X(), &rayMin, &rayMax, boxMin.X()-kEpsilon, boxMax.X()+kEpsilon) {
		return false
	}

	if aabb.intersectsRayInRangeComponent(ray.origin.Y(), ray.direction.Y(), &rayMin, &rayMax, boxMin.Y()-kEpsilon, boxMax.Y()+kEpsilon) {
		return false
	}

	if aabb.intersectsRayInRangeComponent(ray.origin.Z(), ray.direction.Z(), &rayMin, &rayMax, boxMin.Z()-kEpsilon, boxMax.Z()+kEpsilon) {
		return false
	}
	return true
}

func (aabb AABB) intersectsRayInRangeComponent(origin, dir float64, t_min, t_max *float64, boxMin, boxMax float64) bool {
	invD := 1.0 / dir
	t0 := (boxMin - origin) * invD
	t1 := (boxMax - origin) * invD
	if t1 < t0 {
		// std::swap(t0, t1);
		t1, t0 = t0, t1
	}

	if t0 > *t_min {
		*t_min = t0
	}

	if t1 < *t_max {
		*t_max = t1
	}

	return *t_max <= *t_min
}
