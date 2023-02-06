package rendering

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Sphere struct {
	radius    float64
	mat       Material
	animation sample.FloatToVec3
}

func NewSphere(center vector3.Float64, radius float64, mat Material) *Sphere {
	return &Sphere{
		radius:    radius,
		mat:       mat,
		animation: func(t float64) vector3.Float64 { return center },
	}
}

func NewAnimatedSphere(radius float64, mat Material, animation sample.FloatToVec3) *Sphere {
	if animation == nil {
		panic("sphere animation can not be nil")
	}
	return &Sphere{
		radius:    radius,
		mat:       mat,
		animation: animation,
	}
}

func (s Sphere) GetMaterial() Material {
	return s.mat
}

func (s Sphere) UV(p vector3.Float64) vector2.Float64 {
	theta := math.Acos(-p.Y())
	phi := math.Atan2(-p.Z(), p.X()) + math.Pi

	return vector2.New(
		phi/(2*math.Pi),
		theta/math.Pi,
	)
}

func (s Sphere) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	center := s.animation(ray.time)

	oc := ray.Origin().Sub(center)
	a := ray.Direction().Dot(ray.Direction())
	halfB := oc.Dot(ray.Direction())
	c := oc.Dot(oc) - (s.radius * s.radius)

	discriminant := (halfB * halfB) - (a * c)
	if discriminant < 0 {
		return false
	}
	sqrtd := math.Sqrt(discriminant)

	root := (-halfB - sqrtd) / a
	if root < minDistance || maxDistance < root {
		root = (-halfB + sqrtd) / a
		if root < minDistance || maxDistance < root {
			return false
		}
	}

	hitRecord.Distance = root
	hitRecord.Point = ray.At(hitRecord.Distance)
	hitRecord.Normal = hitRecord.Point.Sub(center).DivByConstant(s.radius)
	hitRecord.Material = s.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	hitRecord.UV = s.UV(hitRecord.Normal)

	return true
}

func (s Sphere) BoundingBox(startTime, endTime float64) *geometry.AABB {
	boxSize := vector3.One[float64]().Scale(s.radius)
	// TODO: Need a smarter method... This doesn't work for non-linear lines.
	bs := geometry.NewAABB(s.animation(startTime), boxSize)
	be := geometry.NewAABB(s.animation(endTime), boxSize)

	bs.EncapsulateBounds(be)
	return &bs
}
