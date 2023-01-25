package modeling

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
)

type AABB struct {
	center  vector3.Float64
	extents vector3.Float64
}

func NewAABB(center, size vector3.Float64) AABB {
	return AABB{
		center:  center,
		extents: size.Scale(0.5),
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

	center := max.
		Sub(min).
		DivByConstant(2).
		Add(min)

	return NewAABB(center, max.Sub(min))
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

func (aabb *AABB) Expand(amount float64) {
	a := amount * 0.5
	aabb.extents = aabb.extents.Add(vector3.New(a, a, a))
}

func (aabb AABB) Contains(p vector3.Float64) bool {
	min := aabb.Min()
	max := aabb.Max()

	if p.X() < min.X() {
		return false
	}
	if p.Y() < min.Y() {
		return false
	}
	if p.Z() < min.Z() {
		return false
	}
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

func (aabb *AABB) EncapsulateTri(t Tri) {
	aabb.EncapsulatePoint(t.P1Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P2Vec3Attr(PositionAttribute))
	aabb.EncapsulatePoint(t.P3Vec3Attr(PositionAttribute))
}

func (aabb *AABB) EncapsulateBounds(b AABB) {
	aabb.EncapsulatePoint(b.center.Sub(b.extents))
	aabb.EncapsulatePoint(b.center.Add(b.extents))
}

func (aabb AABB) ClosestPoint(v vector3.Float64) vector3.Float64 {
	result := v
	min := aabb.Min()
	max := aabb.Max()
	result = result.SetX(Clamp(v.X(), min.X(), max.X()))
	result = result.SetY(Clamp(v.Y(), min.Y(), max.Y()))
	result = result.SetZ(Clamp(v.Z(), min.Z(), max.Z()))
	return result
}
