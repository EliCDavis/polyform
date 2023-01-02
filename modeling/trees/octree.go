package trees

import (
	"log"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

type OctTree struct {
	children []*OctTree
	tris     []modeling.Tri
	bounds   modeling.AABB
}

func (ot OctTree) ClosestPoint(v vector.Vector3) vector.Vector3 {
	closestTriDist := math.MaxFloat64
	closestTriPoint := vector.Vector3Zero()

	for i := 0; i < len(ot.tris); i++ {
		point := ot.tris[i].ClosestPoint(v)
		dist := point.Distance(v)
		if dist < closestTriDist {
			closestTriDist = dist
			closestTriPoint = point
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

	if closestCell != nil && closestCellDist < closestTriDist {
		subCellPoint := closestCell.ClosestPoint(v)
		subCellDist := v.Distance(subCellPoint)
		if subCellDist < closestTriDist {
			return subCellPoint
		}
	}

	return closestTriPoint
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

func fromMesh(tris []modeling.Tri, maxDepth int) *OctTree {

	if len(tris) == 0 {
		return nil
	}

	if len(tris) == 1 {
		return &OctTree{
			bounds:   tris[0].Bounds(),
			tris:     []modeling.Tri{tris[0]},
			children: nil,
		}
	}

	bounds := tris[0].Bounds()

	for _, item := range tris {
		bounds.EncapsulateTri(item)
	}

	if maxDepth == 0 {
		return &OctTree{
			bounds:   bounds,
			tris:     tris,
			children: nil,
		}
	}

	var childrenNodes = [][]modeling.Tri{
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
		make([]modeling.Tri, 0),
	}

	globalCenter := bounds.Center()
	leftOver := make([]modeling.Tri, 0)
	for _, item := range tris {
		triBounds := item.Bounds()
		var minIndex = octreeIndex(globalCenter, triBounds.Min())
		var maxIndex = octreeIndex(globalCenter, triBounds.Max())

		if minIndex == maxIndex {
			// child is contained completely within the division, pass it down.
			childrenNodes[minIndex] = append(childrenNodes[minIndex], item)
		} else {
			// Doesn't fit within a single subdivision, stop recursing for this item.
			leftOver = append(leftOver, item)
		}
	}

	return &OctTree{
		bounds: bounds,
		tris:   leftOver,
		children: []*OctTree{
			fromMesh(childrenNodes[0], maxDepth-1),
			fromMesh(childrenNodes[1], maxDepth-1),
			fromMesh(childrenNodes[2], maxDepth-1),
			fromMesh(childrenNodes[3], maxDepth-1),
			fromMesh(childrenNodes[4], maxDepth-1),
			fromMesh(childrenNodes[5], maxDepth-1),
			fromMesh(childrenNodes[6], maxDepth-1),
			fromMesh(childrenNodes[7], maxDepth-1),
		},
	}
}

func FromMeshWithDepth(m modeling.Mesh, depth int) *OctTree {
	tris := make([]modeling.Tri, m.TriCount())
	for i := 0; i < m.TriCount(); i++ {
		tris[i] = m.Tri(i)
	}

	return fromMesh(tris, depth)
}

func logBase8(x float64) float64 {
	return math.Log(x) / math.Log(8)
}

func FromMesh(m modeling.Mesh) *OctTree {
	treeDepth := int(math.Max(1, math.Round(logBase8(float64(m.TriCount())))))
	log.Printf("tree of depth: %d\n", treeDepth)
	return FromMeshWithDepth(m, treeDepth)
}
