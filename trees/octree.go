package trees

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type primitiveReference struct {
	primitive     Element
	originalIndex int
}

type OctTree struct {
	children   []*OctTree
	primitives []primitiveReference
	bounds     geometry.AABB
}

func (ot OctTree) BoundingBox() geometry.AABB {
	return ot.bounds
}

func (ot OctTree) ElementsIntersectingRay(ray geometry.Ray, min, max float64) []int {
	intersections := make([]int, 0)

	for i := 0; i < len(ot.primitives); i++ {
		if ot.primitives[i].primitive.BoundingBox().IntersectsRayInRange(ray, min, max) {
			intersections = append(intersections, ot.primitives[i].originalIndex)
		}
	}

	for _, subtree := range ot.children {
		if subtree != nil && subtree.bounds.IntersectsRayInRange(ray, min, max) {
			intersections = append(intersections, subtree.ElementsIntersectingRay(ray, min, max)...)
		}
	}

	return intersections
}

func (ot OctTree) ElementsContainingPoint(v vector3.Float64) []int {
	intersections := make([]int, 0)

	for i := 0; i < len(ot.primitives); i++ {
		if ot.primitives[i].primitive.BoundingBox().Contains(v) {
			intersections = append(intersections, ot.primitives[i].originalIndex)
		}
	}

	subTreeIndex := octreeIndex(ot.bounds.Center(), v)
	for ot.children[subTreeIndex] != nil {
		return append(intersections, ot.children[subTreeIndex].ElementsContainingPoint(v)...)
	}

	return intersections
}

func (ot OctTree) ClosestPoint(v vector3.Float64) (int, vector3.Float64) {
	closestPrimDist := math.MaxFloat64
	closestPrimPoint := vector3.Zero[float64]()
	closestPointIndex := -1

	for i := 0; i < len(ot.primitives); i++ {
		point := ot.primitives[i].primitive.ClosestPoint(v)
		dist := point.DistanceSquared(v)
		if dist < closestPrimDist {
			closestPrimDist = dist
			closestPrimPoint = point
			closestPointIndex = i
		}
	}

	var closestCell *OctTree = nil
	closestCellDist := math.MaxFloat64

	for i := 0; i < len(ot.children); i++ {
		if ot.children[i] == nil {
			continue
		}
		point := ot.children[i].bounds.ClosestPoint(v)
		dist := point.DistanceSquared(v)
		if dist < closestCellDist {
			closestCellDist = dist
			closestCell = ot.children[i]
		}
	}

	if closestCell != nil && closestCellDist < closestPrimDist {
		cellIndex, subCellPoint := closestCell.ClosestPoint(v)
		subCellDist := v.DistanceSquared(subCellPoint)
		if subCellDist < closestPrimDist {
			return cellIndex, subCellPoint
		}
	}

	return closestPointIndex, closestPrimPoint
}

func octreeIndex(center, item vector3.Float64) int {
	left := 0
	if item.X() < center.X() {
		left = 1
	}

	bottom := 0
	if item.Y() < center.Y() {
		bottom = 2
	}

	back := 0
	if item.Z() < center.Z() {
		back = 4
	}

	return left | bottom | back
}

func newOctree(primitives []primitiveReference, maxDepth int) *OctTree {
	if len(primitives) == 0 {
		return nil
	}

	if len(primitives) == 1 {
		return &OctTree{
			bounds:     primitives[0].primitive.BoundingBox(),
			primitives: []primitiveReference{primitives[0]},
			children:   nil,
		}
	}

	bounds := primitives[0].primitive.BoundingBox()

	for _, item := range primitives {
		bounds.EncapsulateBounds(item.primitive.BoundingBox())
	}

	if maxDepth == 0 {
		return &OctTree{
			bounds:     bounds,
			primitives: primitives,
			children:   nil,
		}
	}

	childrenNodes := [][]primitiveReference{
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
		make([]primitiveReference, 0),
	}

	globalCenter := bounds.Center()
	leftOver := make([]primitiveReference, 0)
	for _, item := range primitives {
		primBounds := item.primitive.BoundingBox()
		minIndex := octreeIndex(globalCenter, primBounds.Min())
		maxIndex := octreeIndex(globalCenter, primBounds.Max())

		if minIndex == maxIndex {
			// child is contained completely within the division, pass it down.
			childrenNodes[minIndex] = append(childrenNodes[minIndex], item)
		} else {
			// Doesn't fit within a single subdivision, stop recursing for this item.
			leftOver = append(leftOver, item)
		}
	}

	return &OctTree{
		bounds:     bounds,
		primitives: leftOver,
		children: []*OctTree{
			newOctree(childrenNodes[0], maxDepth-1),
			newOctree(childrenNodes[1], maxDepth-1),
			newOctree(childrenNodes[2], maxDepth-1),
			newOctree(childrenNodes[3], maxDepth-1),
			newOctree(childrenNodes[4], maxDepth-1),
			newOctree(childrenNodes[5], maxDepth-1),
			newOctree(childrenNodes[6], maxDepth-1),
			newOctree(childrenNodes[7], maxDepth-1),
		},
	}
}

func logBase8(x float64) float64 {
	return math.Log(x) / math.Log(8)
}

func OctreeDepthFromCount(count int) int {
	return int(math.Max(1, math.Round(logBase8(float64(count)))))
}

func NewOctree(elements []Element) *OctTree {
	treeDepth := OctreeDepthFromCount(len(elements))
	return NewOctreeWithDepth(elements, treeDepth)
}

func NewOctreeWithDepth(elements []Element, maxDepth int) *OctTree {
	primitives := make([]primitiveReference, len(elements))
	for i, ele := range elements {
		primitives[i] = primitiveReference{
			primitive:     ele,
			originalIndex: i,
		}
	}
	return newOctree(primitives, maxDepth)
}
