package rendering

import (
	"math/rand"
	"sort"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
)

type axis int

const (
	xAxis axis = iota
	yAxis
	zAxis
	none
)

func nextkdAxis(current axis) axis {
	switch current {
	case xAxis:
		return yAxis
	case yAxis:
		return zAxis
	case zAxis:
		return xAxis
	}
	panic("unimplemented axis")
}

func randomAxis() axis {
	return axis(rand.Intn(3))
}

type BVHNode struct {
	box   geometry.AABB
	left  Hittable
	right Hittable
}

func NewBVHFromMesh(mesh modeling.Mesh, mat Material) *BVHNode {
	its := make([]Hittable, mesh.PrimitiveCount())
	for i := 0; i < mesh.PrimitiveCount(); i++ {
		tri := mesh.Tri(i)
		its[i] = Triangle{
			mat: mat,
			box: geometry.NewAABBFromPoints(
				tri.P1Vec3Attr(modeling.PositionAttribute),
				tri.P2Vec3Attr(modeling.PositionAttribute),
				tri.P3Vec3Attr(modeling.PositionAttribute),
			),

			p1: tri.P1Vec3Attr(modeling.PositionAttribute),
			p2: tri.P2Vec3Attr(modeling.PositionAttribute),
			p3: tri.P3Vec3Attr(modeling.PositionAttribute),

			n1: tri.P1Vec3Attr(modeling.NormalAttribute),
			n2: tri.P2Vec3Attr(modeling.NormalAttribute),
			n3: tri.P3Vec3Attr(modeling.NormalAttribute),
		}
	}

	return NewBVHTree(its, 0, len(its), 0, 0)
}

func NewBVHTree(objects []Hittable, start, end int, startTime, endTime float64) *BVHNode {
	node := &BVHNode{}

	axis := randomAxis()
	objectSpan := end - start

	var comparator sort.Interface = nil
	switch axis {
	case xAxis:
		comparator = SortBoxByXAxis(objects[start:end])

	case yAxis:
		comparator = SortBoxByYAxis(objects[start:end])

	case zAxis:
		comparator = SortBoxByZAxis(objects[start:end])
	}

	if objectSpan == 1 {
		node.left = objects[start]
		node.right = objects[start]
	} else if objectSpan == 2 {
		if comparator.Less(0, 1) {
			node.left = objects[start]
			node.right = objects[start+1]
		} else {
			node.left = objects[start+1]
			node.right = objects[start]
		}
	} else {
		sort.Sort(comparator)
		mid := start + (objectSpan / 2)
		node.left = NewBVHTree(objects, start, mid, startTime, endTime)
		node.right = NewBVHTree(objects, mid, end, startTime, endTime)
	}

	leftBox := node.left.BoundingBox(startTime, endTime)
	rightBox := node.right.BoundingBox(startTime, endTime)
	if leftBox == nil || rightBox == nil {
		panic("no bounding box for element in bvh construction")
	}

	node.box = geometry.NewEmptyAABB()
	node.box.EncapsulateBounds(*leftBox)
	node.box.EncapsulateBounds(*rightBox)

	return node
}

type SortBoxByXAxis []Hittable

func (a SortBoxByXAxis) Len() int      { return len(a) }
func (a SortBoxByXAxis) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortBoxByXAxis) Less(i, j int) bool {
	return a[i].BoundingBox(0, 0).Min().X() < a[j].BoundingBox(0, 0).Min().X()
}

type SortBoxByYAxis []Hittable

func (a SortBoxByYAxis) Len() int      { return len(a) }
func (a SortBoxByYAxis) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortBoxByYAxis) Less(i, j int) bool {
	return a[i].BoundingBox(0, 0).Min().Y() < a[j].BoundingBox(0, 0).Min().Y()
}

type SortBoxByZAxis []Hittable

func (a SortBoxByZAxis) Len() int      { return len(a) }
func (a SortBoxByZAxis) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortBoxByZAxis) Less(i, j int) bool {
	return a[i].BoundingBox(0, 0).Min().Z() < a[j].BoundingBox(0, 0).Min().Z()
}

func (bvhn BVHNode) Hit(r *TemporalRay, min, max float64, hitRecord *HitRecord) bool {
	if !bvhn.box.IntersectsRayInRange(r.Ray(), min, max) {
		return false
	}
	left := bvhn.left.Hit(r, min, max, hitRecord)
	rT := max
	if left {
		rT = hitRecord.Distance
	}
	right := bvhn.right.Hit(r, min, rT, hitRecord)
	return left || right
}

func (bvhn BVHNode) BoundingBox(startTime, endTime float64) *geometry.AABB {
	return &bvhn.box
}
