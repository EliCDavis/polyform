package rendering

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type XYRectangle struct {
	bottomLeft, topRight vector2.Float64
	depth                float64
	mat                  Material
}

func NewXYRectangle(bottomLeft, topRight vector2.Float64, depth float64, mat Material) XYRectangle {
	return XYRectangle{
		bottomLeft: bottomLeft,
		topRight:   topRight,
		depth:      depth,
		mat:        mat,
	}
}

func (xyr XYRectangle) BoundingBox(startTime, endTime float64) *geometry.AABB {
	bb := geometry.NewAABBFromPoints(
		vector3.New(xyr.bottomLeft.X(), xyr.bottomLeft.Y(), xyr.depth-0.0001),
		vector3.New(xyr.topRight.X(), xyr.topRight.Y(), xyr.depth+0.0001),
	)
	return &bb
}

func (xyr XYRectangle) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	t := (xyr.depth - ray.Origin().Z()) / ray.Direction().Z()
	if t < minDistance || t > maxDistance {
		return false
	}
	x := ray.Origin().X() + t*ray.Direction().X()
	y := ray.Origin().Y() + t*ray.Direction().Y()
	if x < xyr.bottomLeft.X() || x > xyr.topRight.X() || y < xyr.bottomLeft.Y() || y > xyr.topRight.Y() {
		return false
	}

	size := xyr.topRight.Sub(xyr.bottomLeft)

	hitRecord.UV = vector2.New(x, y).Sub(xyr.bottomLeft)
	hitRecord.UV = vector2.New(
		hitRecord.UV.X()/size.X(),
		hitRecord.UV.Y()/size.Y(),
	)
	hitRecord.Distance = t
	outward_normal := vector3.New(0., 0., 1.)
	hitRecord.SetFaceNormal(*ray, outward_normal)
	hitRecord.Material = xyr.mat
	hitRecord.Point = ray.At(t)
	return true
}
