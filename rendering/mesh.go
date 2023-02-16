package rendering

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Moller-Trumbor method
// https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/moller-trumbore-ray-triangle-intersection.html
// https://github.com/scratchapixel/code/blob/main/introduction-acceleration-structure/acceleration.cpp#L299
func rayIntersectsTri(tri intersectingTri, ray geometry.Ray, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	const kEpsilon = 0.000001

	dir := ray.Direction()
	orig := ray.At(minDistance)

	v0v1 := tri.p2.Sub(tri.p1)
	v0v2 := tri.p3.Sub(tri.p1)
	pvec := dir.Cross(v0v2)
	det := v0v1.Dot(pvec)

	// ray and triangle are parallel if det is close to 0
	if math.Abs(det) < kEpsilon {
		return false
	}

	invDet := 1. / det

	tvec := orig.Sub(tri.p1)
	u := tvec.Dot(pvec) * invDet
	if u < 0 || u > 1 {
		return false
	}

	qvec := tvec.Cross(v0v1)
	v := dir.Dot(qvec) * invDet
	if v < 0 || u+v > 1 {
		return false
	}

	tVal := v0v2.Dot(qvec) * invDet

	// Prevents us from bouncing around and dying inside the triangle itself
	if tVal < kEpsilon {
		return false
	}

	if tVal > maxDistance {
		return false
	}

	// u v w
	// u w v
	// v w u
	// v u w
	// w u v

	w := 1. - u - v
	normal := tri.n1.Scale(w).
		Add(tri.n2.Scale(u)).
		Add(tri.n3.Scale(v)).
		Normalized()
	// normal = tri.n1.Add(tri.n2).Add(tri.p3).Scale(1. / 3.).Normalized()
	// normal = tri.p1.Sub(tri.p2).Cross(tri.p3.Sub(tri.p2)).Normalized()

	hitRecord.Normal = normal
	hitRecord.Distance = tVal + minDistance
	hitRecord.Point = ray.At(tVal + minDistance)
	hitRecord.Float3Data["barycentric"] = vector3.New(u, v, w)

	return true
}

type intersectingTri struct {
	p1, p2, p3 vector3.Float64
	n1, n2, n3 vector3.Float64
}

func (it intersectingTri) BoundingBox() geometry.AABB {
	return geometry.NewAABBFromPoints(it.p1, it.p2, it.p3)
}

func (it intersectingTri) ClosestPoint(p vector3.Float64) vector3.Float64 {
	panic("unimplemented")
}

type Mesh struct {
	mesh []intersectingTri
	mat  Material
	tree trees.Tree
}

func NewMesh(mesh modeling.Mesh, mat Material) Mesh {
	its := make([]intersectingTri, mesh.PrimitiveCount())
	eles := make([]trees.Element, mesh.PrimitiveCount())
	for i := 0; i < mesh.PrimitiveCount(); i++ {
		tri := mesh.Tri(i)
		its[i] = intersectingTri{
			p1: tri.P1Vec3Attr(modeling.PositionAttribute),
			p2: tri.P2Vec3Attr(modeling.PositionAttribute),
			p3: tri.P3Vec3Attr(modeling.PositionAttribute),

			n1: tri.P1Vec3Attr(modeling.NormalAttribute),
			n2: tri.P2Vec3Attr(modeling.NormalAttribute),
			n3: tri.P3Vec3Attr(modeling.NormalAttribute),
		}
		eles[i] = its[i]
	}

	return Mesh{
		mesh: its,
		mat:  mat,
		tree: trees.NewOctree(eles),
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

func (s Mesh) Hit2(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	intersections := s.tree.ElementsIntersectingRay(ray.Ray(), minDistance, maxDistance)
	if len(intersections) == 0 {
		return false
	}

	hitAnything := false
	closestSoFar := maxDistance

	geoRay := geometry.NewRay(ray.At(minDistance), ray.Direction())
	geoRay = ray.Ray()

	for _, itemIndex := range intersections {
		tri := s.mesh[itemIndex]
		if rayIntersectsTri(tri, geoRay, minDistance, closestSoFar, hitRecord) {
			hitAnything = true
			closestSoFar = hitRecord.Distance
		}
	}

	if !hitAnything {
		return false
	}

	// hitRecord.Distance = root
	// hitRecord.Point = ray.At(hitRecord.Distance)
	// hitRecord.Normal = hitRecord.Point.Sub(center).DivByConstant(s.radius)
	hitRecord.Material = s.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	hitRecord.UV = s.UV(hitRecord.Normal)

	return hitAnything
}

func (s Mesh) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	minStartDistance := minDistance
	maxStartDistance := maxDistance

	hitAnything := false
	// geoRay := geometry.NewRay(ray.At(minDistance), ray.Direction())
	geoRay := ray.Ray()
	s.tree.TraverseIntersectingRay(geoRay, minStartDistance, maxStartDistance, func(i int, min, max *float64) {
		tri := s.mesh[i]
		if rayIntersectsTri(tri, geoRay, minDistance, maxStartDistance, hitRecord) {
			hitAnything = true
			maxStartDistance = hitRecord.Distance
		}
	})

	if !hitAnything {
		return false
	}

	hitRecord.Material = s.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)
	hitRecord.UV = s.UV(hitRecord.Normal)

	return hitAnything
}

func (m Mesh) BoundingBox(startTime, endTime float64) *geometry.AABB {
	boxSize := m.tree.BoundingBox()
	return &boxSize
}
