package rendering

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/trees"
	"github.com/EliCDavis/vector/vector3"
)

type boxElement struct {
	box geometry.AABB
}

func (be boxElement) BoundingBox() geometry.AABB {
	return be.box
}

func (be boxElement) ClosestPoint(p vector3.Float64) vector3.Float64 {
	return be.box.ClosestPoint(p)
}

type Tree struct {
	tree  trees.Tree
	items []Hittable
}

func NewBVH(items []Hittable, startTime, endTime float64) Tree {
	boxElements := make([]trees.Element, len(items))
	for i, h := range items {
		boxElements[i] = boxElement{*h.BoundingBox(startTime, endTime)}
	}
	return Tree{
		tree:  trees.NewOctree(boxElements),
		items: items,
	}
}

func (bvh Tree) Hit(r *TemporalRay, min, max float64, hitRecord *HitRecord) bool {
	intersections := bvh.tree.ElementsIntersectingRay(r.Ray(), min, max)
	tempRecord := NewHitRecord()
	hitAnything := false
	closestSoFar := max

	for _, itemIndex := range intersections {
		item := bvh.items[itemIndex]
		if item.Hit(r, min, closestSoFar, tempRecord) {
			hitAnything = true
			closestSoFar = tempRecord.Distance

			*hitRecord = *tempRecord
		}
	}
	return hitAnything
}

func (bvh Tree) BoundingBox(startTime, endTime float64) *geometry.AABB {
	box := bvh.tree.BoundingBox()
	return &box
}
