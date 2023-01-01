package modeling

import (
	"math"

	"github.com/EliCDavis/vector"
)

type AABB struct {
	center  vector.Vector3
	extents vector.Vector3
}

func NewAABB(center, size vector.Vector3) AABB {
	return AABB{
		center:  center,
		extents: size.MultByConstant(0.5),
	}
}

func (aabb AABB) Center() vector.Vector3 {
	return aabb.center
}

func (aabb AABB) Min() vector.Vector3 {
	return aabb.center.Sub(aabb.extents)
}

func (aabb AABB) Max() vector.Vector3 {
	return aabb.center.Add(aabb.extents)
}

func (aabb AABB) Size() vector.Vector3 {
	return aabb.extents.MultByConstant(2)
}

func (aabb *AABB) Expand(amount float64) {
	a := amount * 0.5
	aabb.extents = aabb.extents.Add(vector.NewVector3(a, a, a))
}

func (aabb AABB) Contains(p vector.Vector3) bool {
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

func minVector(a, b vector.Vector3) vector.Vector3 {
	return vector.NewVector3(
		math.Min(a.X(), b.X()),
		math.Min(a.Y(), b.Y()),
		math.Min(a.Z(), b.Z()),
	)
}

func maxVector(a, b vector.Vector3) vector.Vector3 {
	return vector.NewVector3(
		math.Max(a.X(), b.X()),
		math.Max(a.Y(), b.Y()),
		math.Max(a.Z(), b.Z()),
	)
}

func (aabb *AABB) SetMinMax(min, max vector.Vector3) {
	aabb.extents = max.Sub(min).MultByConstant(0.5)
	aabb.center = min.Add(aabb.extents)
}

func (aabb *AABB) EncapsulatePoint(p vector.Vector3) {
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

func (aabb AABB) ClosestPoint(v vector.Vector3) vector.Vector3 {
	result := v
	min := aabb.Min()
	max := aabb.Max()
	result = result.SetX(Clamp(v.X(), min.X(), max.X()))
	result = result.SetY(Clamp(v.Y(), min.Y(), max.Y()))
	result = result.SetZ(Clamp(v.Z(), min.Z(), max.Z()))
	return result
}
