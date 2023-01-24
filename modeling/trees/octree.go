package trees

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type primitiveReference struct {
	primitive     modeling.Primitive
	originalIndex int
}

type OctTree struct {
	children   []*OctTree
	primitives []primitiveReference
	bounds     modeling.AABB
	atr        string
}

func (ot OctTree) Intersects(v vector3.Float64) []int {
	intersections := make([]int, 0)

	for i := 0; i < len(ot.primitives); i++ {
		if ot.primitives[i].primitive.BoundingBox(ot.atr).Contains(v) {
			intersections = append(intersections, i)
		}
	}

	subTreeIndex := octreeIndex(ot.bounds.Center(), v)
	for ot.children[subTreeIndex] != nil {
		return append(intersections, ot.children[subTreeIndex].Intersects(v)...)
	}

	return intersections
}

func (ot OctTree) ClosestPoint(v vector3.Float64) (int, vector3.Float64) {
	closestPrimDist := math.MaxFloat64
	closestPrimPoint := vector3.Zero[float64]()
	closestPointIndex := -1

	for i := 0; i < len(ot.primitives); i++ {
		point := ot.primitives[i].primitive.ClosestPoint(ot.atr, v)
		dist := point.Distance(v)
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
		dist := point.Distance(v)
		if dist < closestCellDist {
			closestCellDist = dist
			closestCell = ot.children[i]
		}
	}

	if closestCell != nil && closestCellDist < closestPrimDist {
		cellIndex, subCellPoint := closestCell.ClosestPoint(v)
		subCellDist := v.Distance(subCellPoint)
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

func fromPrimitives(primitives []primitiveReference, atr string, maxDepth int) *OctTree {
	if len(primitives) == 0 {
		return nil
	}

	if len(primitives) == 1 {
		return &OctTree{
			bounds:     primitives[0].primitive.BoundingBox(atr),
			primitives: []primitiveReference{primitives[0]},
			children:   nil,
			atr:        atr,
		}
	}

	bounds := primitives[0].primitive.BoundingBox(atr)

	for _, item := range primitives {
		bounds.EncapsulateBounds(item.primitive.BoundingBox(atr))
	}

	if maxDepth == 0 {
		return &OctTree{
			bounds:     bounds,
			primitives: primitives,
			children:   nil,
			atr:        atr,
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
		primBounds := item.primitive.BoundingBox(atr)
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
			fromPrimitives(childrenNodes[0], atr, maxDepth-1),
			fromPrimitives(childrenNodes[1], atr, maxDepth-1),
			fromPrimitives(childrenNodes[2], atr, maxDepth-1),
			fromPrimitives(childrenNodes[3], atr, maxDepth-1),
			fromPrimitives(childrenNodes[4], atr, maxDepth-1),
			fromPrimitives(childrenNodes[5], atr, maxDepth-1),
			fromPrimitives(childrenNodes[6], atr, maxDepth-1),
			fromPrimitives(childrenNodes[7], atr, maxDepth-1),
		},
		atr: atr,
	}
}

func FromPrimitivesWithDepth(primitives []modeling.Primitive, atr string, depth int) *OctTree {
	pr := make([]primitiveReference, len(primitives))
	for i := 0; i < len(primitives); i++ {
		pr[i] = primitiveReference{
			primitive:     primitives[i],
			originalIndex: i,
		}
	}
	return fromPrimitives(pr, atr, depth)
}

func FromPrimitives(primitives []modeling.Primitive, atr string) *OctTree {
	return FromPrimitivesWithDepth(primitives, atr, depthFromCount(len(primitives)))
}

func FromMeshAttributeWithDepth(m modeling.Mesh, atr string, depth int) *OctTree {
	primitives := make([]primitiveReference, m.PrimitiveCount())

	m.ScanPrimitives(func(i int, p modeling.Primitive) {
		primitives[i] = primitiveReference{
			primitive:     p,
			originalIndex: i,
		}
	})

	return fromPrimitives(primitives, atr, depth)
}

func FromMeshWithDepth(m modeling.Mesh, depth int) *OctTree {
	return FromMeshAttributeWithDepth(m, modeling.PositionAttribute, depth)
}

func logBase8(x float64) float64 {
	return math.Log(x) / math.Log(8)
}

func depthFromCount(count int) int {
	return int(math.Max(1, math.Round(logBase8(float64(count)))))
}

func FromMesh(m modeling.Mesh) *OctTree {
	treeDepth := depthFromCount(m.PrimitiveCount())
	return FromMeshAttributeWithDepth(m, modeling.PositionAttribute, treeDepth)
}
