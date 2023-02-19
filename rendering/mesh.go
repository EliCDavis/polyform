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

	w := 1. - u - v
	// normal := tri.n1.Scale(w).
	// 	Add(tri.n2.Scale(u)).
	// 	Add(tri.n3.Scale(v)).
	// 	Normalized()

	hitRecord.Normal = tri.p1.Sub(tri.p2).Cross(tri.p3.Sub(tri.p2)).Normalized()
	hitRecord.Distance = tVal + minDistance
	hitRecord.Point = ray.At(tVal + minDistance)
	hitRecord.Float3Data["barycentric"] = vector3.New(w, u, v)

	return true
}

type intersectingTri struct {
	p1, p2, p3 vector3.Float64
}

func (it intersectingTri) BoundingBox() geometry.AABB {
	return geometry.NewAABBFromPoints(it.p1, it.p2, it.p3)
}

func (it intersectingTri) ClosestPoint(p vector3.Float64) vector3.Float64 {
	panic("unimplemented")
}

type Mesh struct {
	tris            []intersectingTri
	ancillaryV3Data []map[string]vector3.Float64
	ancillaryV2Data []map[string]vector2.Float64
	mat             Material
	tree            trees.Tree
	v3Atr           []string
	v2Atr           []string
}

func NewMesh(mesh modeling.Mesh, mat Material) Mesh {
	v3Data := make([]string, 0)
	v2Data := make([]string, 0)

	if mesh.HasFloat3Attribute(modeling.NormalAttribute) {
		v3Data = append(v3Data, modeling.NormalAttribute)
	}

	if mesh.HasFloat2Attribute(modeling.TexCoordAttribute) {
		v2Data = append(v2Data, modeling.TexCoordAttribute)
	}

	return NewMeshWithAttributes(mesh, mat, v3Data, v2Data)
}

func NewMeshWithAttributes(mesh modeling.Mesh, mat Material, v3Data, v2Data []string) Mesh {
	its := make([]intersectingTri, mesh.PrimitiveCount())
	eles := make([]trees.Element, mesh.PrimitiveCount())

	ancillaryV3Data := make([]map[string]vector3.Float64, 0)
	ancillaryV2Data := make([]map[string]vector2.Float64, 0)

	for i := 0; i < mesh.PrimitiveCount(); i++ {
		tri := mesh.Tri(i)
		its[i] = intersectingTri{
			p1: tri.P1Vec3Attr(modeling.PositionAttribute),
			p2: tri.P2Vec3Attr(modeling.PositionAttribute),
			p3: tri.P3Vec3Attr(modeling.PositionAttribute),
		}

		ancillaryV3Data = append(
			ancillaryV3Data,
			make(map[string]vector3.Vector[float64]),
			make(map[string]vector3.Vector[float64]),
			make(map[string]vector3.Vector[float64]),
		)

		ancillaryV2Data = append(
			ancillaryV2Data,
			make(map[string]vector2.Vector[float64]),
			make(map[string]vector2.Vector[float64]),
			make(map[string]vector2.Vector[float64]),
		)

		for _, keyword := range v3Data {
			ancillaryV3Data[(i*3)+0][keyword] = tri.P1Vec3Attr(keyword)
			ancillaryV3Data[(i*3)+1][keyword] = tri.P2Vec3Attr(keyword)
			ancillaryV3Data[(i*3)+2][keyword] = tri.P3Vec3Attr(keyword)
		}

		for _, keyword := range v2Data {
			ancillaryV2Data[(i*3)+0][keyword] = tri.P1Vec2Attr(keyword)
			ancillaryV2Data[(i*3)+1][keyword] = tri.P2Vec2Attr(keyword)
			ancillaryV2Data[(i*3)+2][keyword] = tri.P3Vec2Attr(keyword)
		}

		eles[i] = its[i]
	}

	return Mesh{
		tris:            its,
		mat:             mat,
		tree:            trees.NewOctree(eles),
		ancillaryV3Data: ancillaryV3Data,
		ancillaryV2Data: ancillaryV2Data,
		v3Atr:           v3Data,
		v2Atr:           v2Data,
	}
}

func (s Mesh) GetMaterial() Material {
	return s.mat
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
		tri := s.tris[itemIndex]
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
	// hitRecord.UV = s.UV(hitRecord.Normal)

	return hitAnything
}

func (s Mesh) Hit(ray *TemporalRay, minDistance, maxDistance float64, hitRecord *HitRecord) bool {
	minStartDistance := minDistance
	maxStartDistance := maxDistance

	// geoRay := geometry.NewRay(ray.At(minDistance), ray.Direction())
	geoRay := ray.Ray()
	closestTriIndex := -1
	s.tree.TraverseIntersectingRay(geoRay, minStartDistance, maxStartDistance, func(i int, min, max *float64) {
		tri := s.tris[i]
		if rayIntersectsTri(tri, geoRay, minDistance, maxStartDistance, hitRecord) {
			closestTriIndex = i
			maxStartDistance = hitRecord.Distance
		}
	})

	if closestTriIndex == -1 {
		return false
	}

	barycentric := hitRecord.Float3Data["barycentric"]

	v3P1Data := s.ancillaryV3Data[(closestTriIndex*3)+0]
	v3P2Data := s.ancillaryV3Data[(closestTriIndex*3)+1]
	v3P3Data := s.ancillaryV3Data[(closestTriIndex*3)+2]
	for _, keyword := range s.v3Atr {
		hitRecord.Float3Data[keyword] = v3P1Data[keyword].Scale(barycentric.X()).
			Add(v3P2Data[keyword].Scale(barycentric.Y())).
			Add(v3P3Data[keyword].Scale(barycentric.Z())).
			Normalized()

		if keyword == modeling.NormalAttribute {
			hitRecord.Normal = hitRecord.Float3Data[keyword]
		}
	}

	v2P1Data := s.ancillaryV2Data[(closestTriIndex*3)+0]
	v2P2Data := s.ancillaryV2Data[(closestTriIndex*3)+1]
	v2P3Data := s.ancillaryV2Data[(closestTriIndex*3)+2]
	for _, keyword := range s.v2Atr {
		hitRecord.Float2Data[keyword] = v2P1Data[keyword].Scale(barycentric.X()).
			Add(v2P2Data[keyword].Scale(barycentric.Y())).
			Add(v2P3Data[keyword].Scale(barycentric.Z())).
			Normalized()

		if keyword == modeling.TexCoordAttribute {
			hitRecord.UV = hitRecord.Float2Data[keyword]
		}
	}

	hitRecord.Material = s.mat
	hitRecord.SetFaceNormal(*ray, hitRecord.Normal)

	return true
}

func (m Mesh) BoundingBox(startTime, endTime float64) *geometry.AABB {
	boxSize := m.tree.BoundingBox()
	return &boxSize
}
