package rendering

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
)

type Sphere struct {
	center vector3.Float64
	radius float64
	mat    Material
}

func NewSphere(center vector3.Float64, radius float64, mat Material) *Sphere {
	return &Sphere{center, radius, mat}
}

func (s Sphere) GetMaterial() Material {
	return s.mat
}

func (s *Sphere) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	oc := ray.Origin().Sub(s.center)
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
	hitRecord.Normal = hitRecord.Point.Sub(s.center).DivByConstant(s.radius)
	hitRecord.Material = s.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)

	return true
}
