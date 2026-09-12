package trees

import (
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

type slabRay struct {
	origin, inverse vector3.Float64
}

func newSlabRay(ray geometry.Ray) slabRay {
	return slabRay{origin: ray.Origin(), inverse: ray.Direction().Reciprocal()}
}

func (r slabRay) hits(box geometry.AABB, tMin, tMax float64) bool {
	const kEpsilon = 0.0000000001
	low, high := box.Min(), box.Max()

	if slab(low.X()-kEpsilon, high.X()+kEpsilon, r.origin.X(), r.inverse.X(), &tMin, &tMax) {
		return false
	}
	if slab(low.Y()-kEpsilon, high.Y()+kEpsilon, r.origin.Y(), r.inverse.Y(), &tMin, &tMax) {
		return false
	}
	return !slab(low.Z()-kEpsilon, high.Z()+kEpsilon, r.origin.Z(), r.inverse.Z(), &tMin, &tMax)
}

func slab(low, high, origin, inverse float64, tMin, tMax *float64) bool {
	t0 := (low - origin) * inverse
	t1 := (high - origin) * inverse
	if t1 < t0 {
		t0, t1 = t1, t0
	}
	if t0 > *tMin {
		*tMin = t0
	}
	if t1 < *tMax {
		*tMax = t1
	}
	return *tMax <= *tMin
}

func (ot *OctTree) ElementsIntersectingRay(ray geometry.Ray, min, max float64) []int {
	ot.intersectionsBuffer = ot.collectRay(newSlabRay(ray), min, max, ot.intersectionsBuffer[:0])
	return ot.intersectionsBuffer
}

func (ot *OctTree) collectRay(ray slabRay, min, max float64, out []int) []int {
	if !ray.hits(ot.bounds, min, max) {
		return out
	}

	for i := range ot.elements {
		if ray.hits(ot.elements[i].bounds, min, max) {
			out = append(out, ot.elements[i].originalIndex)
		}
	}

	for _, child := range ot.children {
		out = child.collectRay(ray, min, max, out)
	}
	return out
}

func (ot *OctTree) TraverseIntersectingRay(ray geometry.Ray, min, max float64, iterator func(i int, min, max *float64)) {
	ot.traverseRay(newSlabRay(ray), &min, &max, iterator)
}

func (ot *OctTree) traverseRay(ray slabRay, min, max *float64, iterator func(i int, min, max *float64)) {
	if !ray.hits(ot.bounds, *min, *max) {
		return
	}

	for i := range ot.elements {
		if ray.hits(ot.elements[i].bounds, *min, *max) {
			iterator(ot.elements[i].originalIndex, min, max)
		}
	}

	for _, child := range ot.children {
		child.traverseRay(ray, min, max, iterator)
	}
}

func (ot *OctTree) ElementsContainingPoint(v vector3.Float64) []int {
	return ot.collectContaining(v, make([]int, 0))
}

func (ot *OctTree) collectContaining(v vector3.Float64, out []int) []int {
	if !ot.bounds.Contains(v) {
		return out
	}

	for i := range ot.elements {
		if ot.elements[i].bounds.Contains(v) {
			out = append(out, ot.elements[i].originalIndex)
		}
	}

	for _, child := range ot.children {
		out = child.collectContaining(v, out)
	}
	return out
}

func (ot *OctTree) ElementsWithinRange(position vector3.Float64, distance float64) []int {
	return ot.collectWithin(position, distance*distance, make([]int, 0))
}

func (ot *OctTree) collectWithin(position vector3.Float64, distanceSquared float64, out []int) []int {
	if ot.bounds.ClosestPoint(position).DistanceSquared(position) > distanceSquared {
		return out
	}

	for i := range ot.elements {
		if ot.elements[i].bounds.ClosestPoint(position).DistanceSquared(position) <= distanceSquared {
			out = append(out, ot.elements[i].originalIndex)
		}
	}

	for _, child := range ot.children {
		out = child.collectWithin(position, distanceSquared, out)
	}
	return out
}

type octDistItem struct {
	dist  float64
	cell  *OctTree
	index int
	point vector3.Float64
}

type octDistHeap []octDistItem

func (h *octDistHeap) push(item octDistItem) {
	*h = append(*h, item)
	items := *h
	i := len(items) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if items[parent].dist <= items[i].dist {
			break
		}
		items[parent], items[i] = items[i], items[parent]
		i = parent
	}
}

func (h *octDistHeap) pop() octDistItem {
	items := *h
	top := items[0]
	last := len(items) - 1
	items[0] = items[last]
	items = items[:last]
	*h = items

	i := 0
	for {
		left, right := 2*i+1, 2*i+2
		smallest := i
		if left < len(items) && items[left].dist < items[smallest].dist {
			smallest = left
		}
		if right < len(items) && items[right].dist < items[smallest].dist {
			smallest = right
		}
		if smallest == i {
			return top
		}
		items[i], items[smallest] = items[smallest], items[i]
		i = smallest
	}
}

func (ot *OctTree) ClosestPoint(v vector3.Float64) (int, vector3.Float64) {
	queue := make(octDistHeap, 0, 64)
	queue.push(octDistItem{
		dist: ot.bounds.ClosestPoint(v).DistanceSquared(v),
		cell: ot,
	})

	for len(queue) > 0 {
		item := queue.pop()

		if item.cell == nil {
			return item.index, item.point
		}

		for _, child := range item.cell.children {
			queue.push(octDistItem{
				dist: child.bounds.ClosestPoint(v).DistanceSquared(v),
				cell: child,
			})
		}
		for i := range item.cell.elements {
			point := item.cell.elements[i].primitive.ClosestPoint(v)
			queue.push(octDistItem{
				dist:  point.DistanceSquared(v),
				index: item.cell.elements[i].originalIndex,
				point: point,
			})
		}
	}

	return -1, vector3.Zero[float64]()
}

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

func octantOf(center vector3.Float64, bounds geometry.AABB) int {
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

const retainedOctant = 8

func newOctree(elements, scratch []elementReference, elementAssignedOctant []uint8, maxDepth int, tolerance float64) *OctTree {
	if len(elements) == 0 {
		return nil
	}

	if len(elements) == 1 {
		return &OctTree{bounds: elements[0].bounds, elements: elements}
	}

	bounds := boundsOf(elements)

	if maxDepth == 0 {
		return &OctTree{bounds: bounds, elements: elements}
	}

	globalCenter := bounds.Center()
	nodeSize := bounds.Size()

	var bucketSizes [9]int
	for i := 0; i < len(elements); i++ {
		bucket := retainedOctant
		if !spansNode(nodeSize, elements[i].bounds.Size(), tolerance) {
			bucket = octantOf(globalCenter, elements[i].bounds)
		}
		elementAssignedOctant[i] = uint8(bucket)
		bucketSizes[bucket]++
	}

	var bucketOffset [9]int
	currentOffset := 0
	for i, size := range bucketSizes {
		bucketOffset[i] = currentOffset
		currentOffset += size
	}

	for i := 0; i < len(elements); i++ {
		octant := elementAssignedOctant[i]
		scratch[bucketOffset[octant]] = elements[i]
		bucketOffset[octant]++
	}
	copy(elements, scratch)

	nonEmptyOctants := 0
	for _, count := range bucketSizes[:retainedOctant] {
		if count > 0 {
			nonEmptyOctants++
		}
	}

	children := make([]*OctTree, 0, nonEmptyOctants)
	first := 0
	for _, octantSize := range bucketSizes[:retainedOctant] {
		last := first + octantSize
		if octantSize > 0 {
			children = append(children, newOctree(
				elements[first:last:last],
				scratch[first:last:last],
				elementAssignedOctant[first:last:last],
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
		bounds:   bounds,
		elements: retained,
		children: children,
	}
}

func logBase8(x float64) float64 {
	return math.Log(x) / math.Log(8)
}

func OctreeDepthFromCount(count int) int {
	return int(math.Max(1, math.Round(logBase8(float64(count)))))
}

const noRetention = -1.

func NewOctree(elements []Element, settings ...OctreeConsturctionSetting) *OctTree {

	construction := &OctreeConsturctionSettings{
		elements:  elements,
		maxDepth:  OctreeDepthFromCount(len(elements)),
		tolerance: 0.2,
	}
	for _, s := range settings {
		s.Configure(construction)
	}

	primitives := make([]elementReference, len(elements))
	for i := range elements {
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
		construction.maxDepth,
		construction.tolerance,
	)
}
