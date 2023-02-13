package rendering

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type intersectingTri struct {
	p1, p2, p3 vector3.Float64
	n1, n2, n3 vector3.Float64
}

func (it intersectingTri) BoundingBox() geometry.AABB {
	return geometry.NewAABBFromPoints(it.p1, it.p2, it.p3)
}

// https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/moller-trumbore-ray-triangle-intersection.html
func rayIntersectsTri(p1, p2, p3 vector3.Float64, ray TemporalRay) (vector3.Float64, bool) {
	const kEpsilon = 0.00001

	dir := ray.Direction()
	orig := ray.Origin()

	v0v1 := p2.Sub(p1)
	v0v2 := p3.Sub(p1)
	pvec := dir.Cross(v0v2)
	det := v0v1.Dot(pvec)

	// ray and triangle are parallel if det is close to 0
	if math.Abs(det) < kEpsilon {
		return vector3.Zero[float64](), false
	}

	invDet := 1. / det

	tvec := orig.Sub(p1)
	u := tvec.Dot(pvec) * invDet
	if u < 0 || u > 1 {
		return vector3.Zero[float64](), false
	}

	qvec := tvec.Cross(v0v1)
	v := dir.Dot(qvec) * invDet
	if v < 0 || u+v > 1 {
		return vector3.Zero[float64](), false
	}

	tVal := v0v2.Dot(qvec) * invDet

	return ray.At(tVal), true
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

func triHit(tri intersectingTri, ray TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	// point, interests := tri.LineIntersects(geometry.NewLine3D(ray.At(minDistance), ray.At(maxDistance)))
	// if !interests {
	// 	return false
	// }

	// if point.Distance(ray.At(0)) < 0.001 {
	// 	return false
	// }

	point, intersects := rayIntersectsTri(tri.p1, tri.p2, tri.p3, ray)
	if !intersects {
		return false
	}

	distFromOrig := point.Distance(ray.Origin())

	// Prevents us from bouncing around and dying inside the triangle itself
	if distFromOrig < 0.001 {
		return false
	}

	if distFromOrig > maxDistance {
		return false
	}

	// compute the plane's normal
	v0v1 := tri.p2.Sub(tri.p1)
	v0v2 := tri.p3.Sub(tri.p1)
	// no need to normalize
	N := v0v1.Cross(v0v2) // N
	denom := N.Dot(N)

	edge1 := tri.p3.Sub(tri.p2)
	vp1 := point.Sub(tri.p2)
	C := edge1.Cross(vp1)
	u := N.Dot(C)

	edge2 := tri.p1.Sub(tri.p3)
	vp2 := point.Sub(tri.p3)
	C = edge2.Cross(vp2)
	v := N.Dot(C)

	u /= denom
	v /= denom

	w := 1. - u - v
	normal := tri.n1.Scale(u).
		Add(tri.n2.Scale(v)).
		Add(tri.n3.Scale(w)).
		Normalized()

	hitRecord.Normal = normal
	hitRecord.Distance = distFromOrig
	hitRecord.Point = point
	hitRecord.Float3Data["barycentric"] = vector3.New(u, v, w)

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

	// geoRay := geometry.NewRay(ray.At(minDistance), ray.Direction())

	for _, itemIndex := range intersections {
		tri := s.mesh[itemIndex]
		if triHit(tri, *ray, minDistance, closestSoFar, tempRecord) {
			hitAnything = true
			closestSoFar = tempRecord.Distance
		}
	}

	if !hitAnything {
		return false
	}

	*hitRecord = *tempRecord

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
