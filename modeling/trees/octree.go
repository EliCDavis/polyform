package trees

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

type OctTree struct {
	children   []*OctTree
	primitives []modeling.Primitive
	bounds     modeling.AABB
	atr        string
}

func (ot OctTree) ClosestPoint(v vector.Vector3) vector.Vector3 {
	closestPrimDist := math.MaxFloat64
	closestPrimPoint := vector.Vector3Zero()

	for i := 0; i < len(ot.primitives); i++ {
		point := ot.primitives[i].ClosestPoint(ot.atr, v)
		dist := point.Distance(v)
		if dist < closestPrimDist {
			closestPrimDist = dist
			closestPrimPoint = point
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
		subCellPoint := closestCell.ClosestPoint(v)
		subCellDist := v.Distance(subCellPoint)
		if subCellDist < closestPrimDist {
			return subCellPoint
		}
	}

	return closestPrimPoint
}

func octreeIndex(center, item vector.Vector3) int {
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

func fromMesh(primitives []modeling.Primitive, atr string, maxDepth int) *OctTree {

	if len(primitives) == 0 {
		return nil
	}

	if len(primitives) == 1 {
		return &OctTree{
			bounds:     primitives[0].BoundingBox(atr),
			primitives: []modeling.Primitive{primitives[0]},
			children:   nil,
			atr:        atr,
		}
	}

	bounds := primitives[0].BoundingBox(atr)

	for _, item := range primitives {
		bounds.EncapsulateBounds(item.BoundingBox(atr))
	}

	if maxDepth == 0 {
		return &OctTree{
			bounds:     bounds,
			primitives: primitives,
			children:   nil,
			atr:        atr,
		}
	}

	var childrenNodes = [][]modeling.Primitive{
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
		make([]modeling.Primitive, 0),
	}

	globalCenter := bounds.Center()
	leftOver := make([]modeling.Primitive, 0)
	for _, item := range primitives {
		primBounds := item.BoundingBox(atr)
		var minIndex = octreeIndex(globalCenter, primBounds.Min())
		var maxIndex = octreeIndex(globalCenter, primBounds.Max())

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
			fromMesh(childrenNodes[0], atr, maxDepth-1),
			fromMesh(childrenNodes[1], atr, maxDepth-1),
			fromMesh(childrenNodes[2], atr, maxDepth-1),
			fromMesh(childrenNodes[3], atr, maxDepth-1),
			fromMesh(childrenNodes[4], atr, maxDepth-1),
			fromMesh(childrenNodes[5], atr, maxDepth-1),
			fromMesh(childrenNodes[6], atr, maxDepth-1),
			fromMesh(childrenNodes[7], atr, maxDepth-1),
		},
		atr: atr,
	}
}

func FromMeshAttributeWithDepth(m modeling.Mesh, atr string, depth int) *OctTree {
	primitives := make([]modeling.Primitive, m.PrimitiveCount())

	m.ScanPrimitives(func(i int, p modeling.Primitive) {
		primitives[i] = p
	})

	return fromMesh(primitives, atr, depth)
}

func FromMeshWithDepth(m modeling.Mesh, depth int) *OctTree {
	return FromMeshAttributeWithDepth(m, modeling.PositionAttribute, depth)
}

func logBase8(x float64) float64 {
	return math.Log(x) / math.Log(8)
}

func FromMesh(m modeling.Mesh) *OctTree {
	treeDepth := int(math.Max(1, math.Round(logBase8(float64(m.PrimitiveCount())))))
	return FromMeshAttributeWithDepth(m, modeling.PositionAttribute, treeDepth)
}
