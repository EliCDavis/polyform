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
		ot.intersectionsBuffer = append(ot.intersectionsBuffer, ot.children[i].ElementsIntersectingRay(ray, min, max)...)
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
		ot.children[i].TraverseIntersectingRay(ray, tMin, tMax, iterator)
	}
}

func (ot OctTree) ElementsContainingPoint(v vector3.Float64) []int {
	intersections := make([]int, 0)

	for i := 0; i < len(ot.elements); i++ {
		if ot.elements[i].bounds.Contains(v) {
			intersections = append(intersections, ot.elements[i].originalIndex)
		}
	}

	for _, child := range ot.children {
		if child.bounds.Contains(v) {
			intersections = append(intersections, child.ElementsContainingPoint(v)...)
		}
	}

	return intersections
}

func (ot OctTree) ElementsWithinRange(position vector3.Float64, distance float64) []int {

	if ot.bounds.ClosestPoint(position).Distance(position) > distance {
		return nil
	}

	points := make([]int, 0)

	for _, ele := range ot.elements {
		if ele.bounds.ClosestPoint(position).Distance(position) <= distance {
			points = append(points, ele.originalIndex)
		}
	}

	for _, child := range ot.children {
		points = append(points, child.ElementsWithinRange(position, distance)...)
	}

	return points
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

// Niave "EncapsulateBounds" is pretty slow if you call it over and over (one per element).
// So do it ourselves in one go.
func boundsOf(elements []elementReference) geometry.AABB {
	low, high := elements[0].bounds.Min(), elements[0].bounds.Max()
	minX, minY, minZ := low.X(), low.Y(), low.Z()
	maxX, maxY, maxZ := high.X(), high.Y(), high.Z()

	for i := 1; i < len(elements); i++ {
		low, high = elements[i].bounds.Min(), elements[i].bounds.Max()
		if low.X() < minX {
			minX = low.X()
		}
		if low.Y() < minY {
			minY = low.Y()
		}
		if low.Z() < minZ {
			minZ = low.Z()
		}
		if high.X() > maxX {
			maxX = high.X()
		}
		if high.Y() > maxY {
			maxY = high.Y()
		}
		if high.Z() > maxZ {
			maxZ = high.Z()
		}
	}

	var bounds geometry.AABB
	bounds.SetMinMax(
		vector3.New(minX, minY, minZ),
		vector3.New(maxX, maxY, maxZ),
	)
	return bounds
}

// Whichever corner sits furthest from the center decides the octant, which
// keeps a straddling element as far from the dividing planes as it can be.
func octantOf(center vector3.Float64, bounds geometry.AABB) int {
	// TODO: See how much faster we can make this
	if center.Sub(bounds.Center()).Dot(bounds.Size()) > 0 {
		return octreeIndex(center, bounds.Min())
	}
	return octreeIndex(center, bounds.Max())
}

func spansNode(nodeSize, elementSize vector3.Float64, tolerance float64) bool {
	fraction := 1 - tolerance
	return spansAxis(nodeSize.X(), elementSize.X(), fraction) ||
		spansAxis(nodeSize.Y(), elementSize.Y(), fraction) ||
		spansAxis(nodeSize.Z(), elementSize.Z(), fraction)
}

// A node flat in an axis has every element flat in it too, so without the zero
// check every element matches on that axis and the whole mesh is retained.
func spansAxis(node, element, fraction float64) bool {
	return node > 0 && element >= node*fraction
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

// The ninth bucket the partition sorts into: the elements this node keeps for
// itself rather than handing to a child.
const retainedOctant = 8

func newOctree(elements, scratch []elementReference, octants []uint8, maxDepth int, tolerance float64) *OctTree {
	if len(elements) == 0 {
		return nil
	}

	if len(elements) == 1 {
		return &OctTree{
			bounds:              elements[0].bounds,
			elements:            elements,
			children:            nil,
			intersectionsBuffer: make([]int, 0),
		}
	}

	bounds := boundsOf(elements)

	if maxDepth == 0 {
		return &OctTree{
			bounds:              bounds,
			elements:            elements,
			children:            nil,
			intersectionsBuffer: make([]int, 0),
		}
	}

	globalCenter := bounds.Center()
	nodeSize := bounds.Size()

	var counts [9]int
	for i := 0; i < len(elements); i++ {
		bucket := retainedOctant
		if !spansNode(nodeSize, elements[i].bounds.Size(), tolerance) {
			bucket = octantOf(globalCenter, elements[i].bounds)
		}
		octants[i] = uint8(bucket)
		counts[bucket]++
	}

	var cursor [9]int
	at := 0
	for i, count := range counts {
		cursor[i] = at
		at += count
	}

	// Scattering in index order keeps elements in their original order within
	// each octant, which several callers rely on for a stable first match.
	for i := 0; i < len(elements); i++ {
		octant := octants[i]
		scratch[cursor[octant]] = elements[i]
		cursor[octant]++
	}
	copy(elements, scratch)

	// newOctree returns nil exactly when handed no elements, so the counts
	// already say how many children there will be.
	found := 0
	for _, count := range counts[:retainedOctant] {
		if count > 0 {
			found++
		}
	}

	children := make([]*OctTree, 0, found)
	first := 0
	for _, count := range counts[:retainedOctant] {
		last := first + count
		if count > 0 {
			// Capped at the run so an append reallocates instead of reaching
			// into the next child's.
			children = append(children, newOctree(
				elements[first:last:last],
				scratch[first:last:last],
				octants[first:last:last],
				maxDepth-1,
				tolerance,
			))
		}
		first = last
	}

	retained := elements[first:]

	if len(children) == 1 && len(retained) == 0 {
		// Prevents us from creating an octree node that's just a proxy to another
		// node. Faster traversal!
		return children[0]
	}

	return &OctTree{
		bounds:              bounds,
		elements:            retained,
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
	return NewOctreeWithRetention(elements, maxDepth, NoRetention)
}

// NoRetention pushes every element down to a leaf.
const NoRetention = -1.

// NewOctreeWithRetention keeps at each node the elements whose bounds come
// within tolerance of matching the node's own size along some axis, rather
// than pushing every element down to a leaf.
//
// An element that wide cannot fit inside any one child, so whichever child
// takes it has its bounds stretched back out over the parent's and stops
// excluding anything. A tolerance of 0 retains only elements exactly as wide
// as the node, 0.5 retains anything at least half as wide, and 1 retains the
// whole mesh at the root. NoRetention turns it off.
func NewOctreeWithRetention(elements []Element, maxDepth int, tolerance float64) *OctTree {
	primitives := make([]elementReference, len(elements))
	for i := 0; i < len(elements); i++ {
		primitives[i] = elementReference{
			primitive:     elements[i],
			originalIndex: i,
			bounds:        elements[i].BoundingBox(),
		}
	}
	return newOctree(
		primitives,
		make([]elementReference, len(primitives)),
		make([]uint8, len(primitives)),
		maxDepth,
		tolerance,
	)
}
