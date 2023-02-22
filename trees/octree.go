package trees

import (
	"container/heap"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
)

type OctTree struct {
	children            []*OctTree
	elements            []elementReference
	bounds              geometry.AABB
	intersectionsBuffer []int
}

func (ot OctTree) BoundingBox() geometry.AABB {
	return ot.bounds
}

func (ot *OctTree) ElementsIntersectingRay(ray geometry.Ray, min, max float64) []int {
	if !ot.bounds.IntersectsRayInRange(ray, min, max) {
		return nil
	}

	ot.intersectionsBuffer = ot.intersectionsBuffer[:0]

	for i := 0; i < len(ot.elements); i++ {
		bounds := ot.elements[i].bounds
		oi := ot.elements[i].originalIndex
		if bounds.IntersectsRayInRange(ray, min, max) {
			ot.intersectionsBuffer = append(ot.intersectionsBuffer, oi)
		}
	}

	for i := 0; i < len(ot.children); i++ {
		if ot.children[i] != nil {
			ot.intersectionsBuffer = append(ot.intersectionsBuffer, ot.children[i].ElementsIntersectingRay(ray, min, max)...)
		}
	}

	return ot.intersectionsBuffer
}

func (ot OctTree) TraverseIntersectingRay(ray geometry.Ray, min, max float64, iterator func(i int, min, max *float64)) {
	if !ot.bounds.IntersectsRayInRange(ray, min, max) {
		return
	}

	tMin := min
	tMax := max

	for i := 0; i < len(ot.elements); i++ {
		bounds := ot.elements[i].bounds
		oi := ot.elements[i].originalIndex
		if bounds.IntersectsRayInRange(ray, tMin, tMax) {
			iterator(oi, &tMin, &tMax)
		}
	}

	for i := 0; i < len(ot.children); i++ {
		if ot.children[i] != nil {
			ot.children[i].TraverseIntersectingRay(ray, tMin, tMax, iterator)
		}
	}
}

func (ot OctTree) ElementsContainingPoint(v vector3.Float64) []int {
	intersections := make([]int, 0)

	for i := 0; i < len(ot.elements); i++ {
		if ot.elements[i].bounds.Contains(v) {
			intersections = append(intersections, ot.elements[i].originalIndex)
		}
	}

	subTreeIndex := octreeIndex(ot.bounds.Center(), v)
	for ot.children[subTreeIndex] != nil {
		return append(intersections, ot.children[subTreeIndex].ElementsContainingPoint(v)...)
	}

	return intersections
}

type octDistItem struct {
	dist    float64
	cell    *OctTree
	element *elementReference
	point   vector3.Float64
}

type octItemPriorityQueue []octDistItem

func (pq octItemPriorityQueue) Len() int { return len(pq) }

func (pq octItemPriorityQueue) Less(i, j int) bool {
	return pq[i].dist < pq[j].dist
}

func (pq octItemPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *octItemPriorityQueue) Push(x any) {
	item := x.(octDistItem)
	*pq = append(*pq, item)
}

func (pq *octItemPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (ot OctTree) ClosestPoint(v vector3.Float64) (int, vector3.Float64) {
	pq := make(octItemPriorityQueue, 1)
	pq[0] = octDistItem{
		dist: ot.bounds.ClosestPoint(v).DistanceSquared(v),
		cell: &ot,
	}

	heap.Init(&pq)

	for pq.Len() > 0 {
		item := heap.Pop(&pq).(octDistItem)

		if item.element != nil {
			return item.element.originalIndex, item.point
		}

		if item.cell != nil {
			for _, child := range item.cell.children {
				if child == nil {
					continue
				}
				heap.Push(&pq, octDistItem{
					dist: child.bounds.ClosestPoint(v).DistanceSquared(v),
					cell: child,
				})
			}
			for _, element := range item.cell.elements {
				point := element.primitive.ClosestPoint(v)

				heap.Push(&pq, octDistItem{
					dist:    point.DistanceSquared(v),
					element: &element,
					point:   point,
				})
			}
		}

	}

	return -1, vector3.Zero[float64]()
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

func newOctree(elements []elementReference, maxDepth int) *OctTree {
	if len(elements) == 0 {
		return nil
	}

	if len(elements) == 1 {
		return &OctTree{
			bounds:              elements[0].bounds,
			elements:            []elementReference{elements[0]},
			children:            nil,
			intersectionsBuffer: make([]int, 0),
		}
	}

	bounds := elements[0].primitive.BoundingBox()

	for _, item := range elements {
		bounds.EncapsulateBounds(item.bounds)
	}

	if maxDepth == 0 {
		return &OctTree{
			bounds:              bounds,
			elements:            elements,
			children:            nil,
			intersectionsBuffer: make([]int, 0),
		}
	}

	childrenNodes := [][]elementReference{
		make([]elementReference, 0),
		make([]elementReference, 0),
		make([]elementReference, 0),
		make([]elementReference, 0),
		make([]elementReference, 0),
		make([]elementReference, 0),
		make([]elementReference, 0),
		make([]elementReference, 0),
	}

	globalCenter := bounds.Center()
	leftOver := make([]elementReference, 0)
	for i := 0; i < len(elements); i++ {
		primBounds := elements[i].bounds
		distMin := globalCenter.Distance(primBounds.Min())
		distMax := globalCenter.Distance(primBounds.Max())

		// Prioritize what will keep us furthest from the center to prevent as
		// much overlap as possible
		if distMin > distMax {
			minIndex := octreeIndex(globalCenter, primBounds.Min())
			childrenNodes[minIndex] = append(childrenNodes[minIndex], elements[i])
		} else {
			maxIndex := octreeIndex(globalCenter, primBounds.Max())
			childrenNodes[maxIndex] = append(childrenNodes[maxIndex], elements[i])
		}

		// minIndex := octreeIndex(globalCenter, primBounds.Min())
		// maxIndex := octreeIndex(globalCenter, primBounds.Max())

		// if minIndex == maxIndex {
		// 	// child is contained completely within the division, pass it down.
		// 	childrenNodes[minIndex] = append(childrenNodes[minIndex], elements[i])
		// } else {
		// 	// Doesn't fit within a single subdivision, stop recursing for this item.
		// 	leftOver = append(leftOver, elements[i])
		// }
	}

	children := []*OctTree{
		newOctree(childrenNodes[0], maxDepth-1),
		newOctree(childrenNodes[1], maxDepth-1),
		newOctree(childrenNodes[2], maxDepth-1),
		newOctree(childrenNodes[3], maxDepth-1),
		newOctree(childrenNodes[4], maxDepth-1),
		newOctree(childrenNodes[5], maxDepth-1),
		newOctree(childrenNodes[6], maxDepth-1),
		newOctree(childrenNodes[7], maxDepth-1),
	}

	if len(leftOver) == 0 {
		var goodChild *OctTree = nil
		goodChildCount := 0
		for _, child := range children {
			if child != nil {
				goodChild = child
				goodChildCount++
			}
		}
		// Prevents us from creating an octree node that's just a proxy to another
		// node. Faster traversal!
		if goodChildCount == 1 {
			return goodChild
		}
	}

	return &OctTree{
		bounds:              bounds,
		elements:            leftOver,
		children:            children,
		intersectionsBuffer: make([]int, 0),
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
	primitives := make([]elementReference, len(elements))
	for i := 0; i < len(elements); i++ {
		primitives[i] = elementReference{
			primitive:     elements[i],
			originalIndex: i,
			bounds:        elements[i].BoundingBox(),
		}
	}
	return newOctree(primitives, maxDepth)
}
