// Package csg does constructive solid geometry on triangle meshes, following
// Laidlaw, Trumbore and Hughes (SIGGRAPH 1986). Section numbers refer to it.
//
//	https://dl.acm.org/doi/10.1145/15922.15904
package csg

import (
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
)

type operation int

const (
	union operation = iota
	intersection
	difference
)

// Where a face sits relative to the other solid. SAME and OPPOSITE mean it
// lies on that solid's boundary, with its normal along or against.
type classification int

const (
	inside classification = iota
	outside
	same
	opposite
)

func (c classification) String() string {
	switch c {
	case inside:
		return "inside"
	case outside:
		return "outside"
	case same:
		return "same"
	}
	return "opposite"
}

// Figure 9.1. A face survives only where its own solid's row says yes. SAME
// and OPPOSITE faces from B are never kept; A's copy already covers them.
type rule map[classification]bool

var keepFromA = map[operation]rule{
	union:        {outside: true, same: true},
	intersection: {inside: true, same: true},
	difference:   {outside: true, opposite: true},
}

var keepFromB = map[operation]rule{
	union:        {outside: true},
	intersection: {inside: true},
	difference:   {inside: true},
}

// Section 9: faces of B inside A are flipped, since B's interior becomes the
// result's exterior.
func flipsFromB(op operation) bool {
	return op == difference
}

// Union returns the mesh enclosing every point in either solid.
func Union(a, b modeling.Mesh) (modeling.Mesh, error) {
	return meshes(a, b, union)
}

// Intersect returns the mesh enclosing the points inside both solids.
func Intersect(a, b modeling.Mesh) (modeling.Mesh, error) {
	return meshes(a, b, intersection)
}

// Subtract returns the mesh enclosing the points of a that lie outside b.
// Coplanar faces are carried over from a unsplit, the paper's known weak spot.
func Subtract(a, b modeling.Mesh) (modeling.Mesh, error) {
	return meshes(a, b, difference)
}

func meshes(a, b modeling.Mesh, op operation) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	first, err := NewSolid(a)
	if err != nil {
		return empty, fmt.Errorf("first %w", err)
	}
	second, err := NewSolid(b)
	if err != nil {
		return empty, fmt.Errorf("second %w", err)
	}
	mesh, _, err := combine(first, second, op)
	if err != nil {
		return empty, err
	}
	return mesh, nil
}

// The pieces come back alongside the mesh, parented to it, so a result can
// become a Solid without reading its own faces back out.
func combine(a, b *Solid, op operation) (modeling.Mesh, []face, error) {
	tolerance := max(a.tolerance, b.tolerance)

	// Section 4. Both are cut before either is split, so each can be split at
	// every point on the shared curve rather than only its own.
	cutsA, curvePointsA, touchingA := cutsAgainst(a.faces, b.faces, tolerance)
	cutsB, curvePointsB, touchingB := cutsAgainst(b.faces, a.faces, tolerance)
	curvePoints := append(curvePointsA, curvePointsB...)

	splitA, err := splitAll(a.faces, a.cornerIDs, cutsA, curvePoints, touchingA, tolerance, newIDSpace(a.weld, tolerance))
	if err != nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil, fmt.Errorf("splitting the first mesh: %w", err)
	}
	splitB, err := splitAll(b.faces, b.cornerIDs, cutsB, curvePoints, touchingB, tolerance, newIDSpace(b.weld, tolerance))
	if err != nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil, fmt.Errorf("splitting the second mesh: %w", err)
	}

	answersA, answersB := classifyBoth(splitA, splitB, tolerance)

	fromA := kept{source: a.mesh, faces: make([]face, 0, len(splitA.faces))}
	fromB := kept{source: b.mesh, faces: make([]face, 0, len(splitB.faces)), inverted: flipsFromB(op)}

	// Section 9, first row of figure 9.1.
	for i, f := range splitA.faces {
		if keepFromA[op][answersA[i]] {
			fromA.faces = append(fromA.faces, f)
		}
	}

	// Section 9, second row.
	for i, f := range splitB.faces {
		if !keepFromB[op][answersB[i]] {
			continue
		}
		if fromB.inverted {
			f = f.reversed()
		}
		fromB.faces = append(fromB.faces, f)
	}

	mesh, faces := meshFromFaces(fromA, fromB)
	return mesh, faces, nil
}

// One solid after splitting: its pieces, their welded corner ids, which lie
// on the other solid's surface, and which ids sit on the shared curve.
type half struct {
	faces     []face
	cornerIDs [][3]int
	touching  []bool
	onCurve   map[int]bool
}

// Section 8 groups faces into regions the intersection curve does not cross
// and settles each with a few rays instead of one per face.
func classifyBoth(splitA, splitB half, tolerance float64) (answersA, answersB []classification) {
	patchesA := patchesOf(splitA.cornerIDs, splitA.onCurve, splitA.touching)
	patchesB := patchesOf(splitB.cornerIDs, splitB.onCurve, splitB.touching)

	// Each solid is cast against by the other's patches, so that is the count
	// that decides whether indexing it pays.
	targetA := newTarget(splitA.faces, tolerance, rayBudget(patchesB))
	targetB := newTarget(splitB.faces, tolerance, rayBudget(patchesA))

	return classifyPatches(splitA.faces, patchesA, targetB),
		classifyPatches(splitB.faces, patchesB, targetA)
}
