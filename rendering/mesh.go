package rendering

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Mesh struct {
	mesh modeling.Mesh
	mat  Material
	tree trees.Tree
}

func NewMesh(mesh modeling.Mesh, mat Material) Mesh {
	return Mesh{
		mesh: mesh,
		mat:  mat,
		tree: mesh.OctTree(),
	}
}

func (s Mesh) GetMaterial() Material {
	return s.mat
}

func (s Mesh) UV(p vector3.Float64) vector2.Float64 {
	theta := math.Acos(-p.Y())
	phi := math.Atan2(-p.Z(), p.X()) + math.Pi

	return vector2.New(
		phi/(2*math.Pi),
		theta/math.Pi,
	)
}

func triHit(tri modeling.Tri, ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	point, interests := tri.LineIntersects(geometry.NewLine3D(ray.At(minDistance), ray.At(maxDistance)))
	if !interests {
		return false
	}

	normal := tri.Average(modeling.NormalAttribute)

	hitRecord.Distance = ray.origin.Distance(point)
	hitRecord.FrontFace = ray.direction.Dot(normal) > 0
	hitRecord.Point = point
	return true
}

func (s Mesh) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	intersections := s.tree.ElementsIntersectingRay(ray.Ray(), minDistance, maxDistance)
	if len(intersections) == 0 {
		return false
	}

	tempRecord := NewHitRecord()
	hitAnything := false
	closestSoFar := maxDistance

	for _, itemIndex := range intersections {
		tri := s.mesh.Tri(itemIndex)
		if triHit(tri, ray, minDistance, closestSoFar, tempRecord) {
			hitAnything = true
			closestSoFar = tempRecord.Distance

			*hitRecord = *tempRecord
		}
	}

	// hitRecord.Distance = root
	// hitRecord.Point = ray.At(hitRecord.Distance)
	// hitRecord.Normal = hitRecord.Point.Sub(center).DivByConstant(s.radius)
	hitRecord.Material = s.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	hitRecord.UV = s.UV(hitRecord.Normal)

	return hitAnything
}

func (m Mesh) BoundingBox(startTime, endTime float64) *geometry.AABB {
	boxSize := m.tree.BoundingBox()
	return &boxSize
}
