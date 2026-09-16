// Package csg performs constructive solid geometry on triangle meshes.
//
// The algorithm is Laidlaw, Trumbore and Hughes, "Constructive Solid Geometry
// for Polyhedral Objects", SIGGRAPH 1986, ACM SIGGRAPH Computer Graphics
// 20(4), pages 161-170.
//
//	https://dl.acm.org/doi/10.1145/15922.15904
//	https://cs.brown.edu/people/jhughes/papers/Laidlaw-CSG-1986/main.htm
//
// Section numbers throughout this package refer to that paper. It runs in
// three stages: both meshes are split along the curve where they meet
// (sections 4 to 6), every resulting face is classified against the opposing
// solid (section 7), and each operation keeps a different set of those
// classifications (section 9).
//
// Two departures from the paper are worth knowing about. Section 6 subdivides
// a polygon by hand, case by case, to keep it convex; this package feeds the
// cut segments to a constrained Delaunay triangulator instead, which reaches
// the same place without the case analysis. And section 4 leaves coplanar
// pairs alone, relying on neighbouring faces to cut them, which this package
// also does - see the note on Subtract.
package csg

import (
	"errors"
	"fmt"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

type operation int

const (
	union operation = iota
	intersection
	difference
)

// Where a face sits relative to the opposing solid, per section 7. SAME and
// OPPOSITE mean the face lies on that solid's boundary, with its normal
// pointing the same way or against.
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

// Figure 9.1. A face survives only where its own solid's row says yes.
//
// The paper never keeps a SAME or OPPOSITE face from B: a face on the shared
// boundary exists in both solids, and the copy from A already covers it.
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

// Section 9: "each polygonB inside objectA must have the order of its
// vertices reversed, and its normal vector must be inverted, since the
// interior of objectB becomes the exterior of the resulting object."
func flipsFromB(op operation) bool {
	return op == difference
}

// Union returns the mesh enclosing every point in either solid.
func Union(a, b modeling.Mesh) (modeling.Mesh, error) {
	return combine(a, b, union)
}

// Intersect returns the mesh enclosing the points inside both solids.
func Intersect(a, b modeling.Mesh) (modeling.Mesh, error) {
	return combine(a, b, intersection)
}

// Subtract returns the mesh enclosing the points of a that lie outside b.
//
// Faces the two solids share exactly are carried over from a, following
// section 4's decision to leave coplanar pairs unsplit. Where such a face
// only partly overlaps and no other face cuts it, the seam there is the
// paper's known weak spot rather than anything this package adds.
func Subtract(a, b modeling.Mesh) (modeling.Mesh, error) {
	return combine(a, b, difference)
}

func combine(a, b modeling.Mesh, op operation) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)

	facesA, err := facesOf(a)
	if err != nil {
		return empty, fmt.Errorf("first mesh: %w", err)
	}
	facesB, err := facesOf(b)
	if err != nil {
		return empty, fmt.Errorf("second mesh: %w", err)
	}
	if len(facesA) == 0 || len(facesB) == 0 {
		return empty, errors.New("both meshes need at least one triangle")
	}

	tolerance := toleranceFor(facesA, facesB)

	// One welder for the whole operation. It pays for itself validating the
	// inputs, and every vertex identity found here is carried through the
	// split rather than rediscovered from coordinates afterwards.
	weldA, weldB := newWelder(tolerance), newWelder(tolerance)
	idsA, idsB := weldFaces(weldA, facesA), weldFaces(weldB, facesB)

	if err := closedFrom(idsA); err != nil {
		return empty, fmt.Errorf("first %w", err)
	}
	if err := closedFrom(idsB); err != nil {
		return empty, fmt.Errorf("second %w", err)
	}

	// Section 4: "The first step in the algorithm is splitting both objects."
	// Both are cut before either is split, so each can be split at every
	// point on the curve they share rather than only its own.
	cutsA, cornersA, touchingA := cutsAgainst(facesA, facesB, tolerance)
	cutsB, cornersB, touchingB := cutsAgainst(facesB, facesA, tolerance)
	corners := append(cornersA, cornersB...)

	splitA, splitIDsA, splitTouchingA, err := splitAll(facesA, idsA, cutsA, corners, touchingA, tolerance, weldA)
	if err != nil {
		return empty, fmt.Errorf("splitting the first mesh: %w", err)
	}
	splitB, splitIDsB, splitTouchingB, err := splitAll(facesB, idsB, cutsB, corners, touchingB, tolerance, weldB)
	if err != nil {
		return empty, fmt.Errorf("splitting the second mesh: %w", err)
	}

	answersA, answersB := classifyBoth(
		splitA, splitIDsA, splitTouchingA, weldA,
		splitB, splitIDsB, splitTouchingB, weldB,
		corners, tolerance,
	)

	kept := make([]face, 0, len(splitA)+len(splitB))

	// Section 9, first row of figure 9.1.
	for i, f := range splitA {
		if keepFromA[op][answersA[i]] {
			kept = append(kept, f)
		}
	}

	// Section 9, second row.
	for i, f := range splitB {
		if !keepFromB[op][answersB[i]] {
			continue
		}
		if flipsFromB(op) {
			f = f.reversed()
		}
		kept = append(kept, f)
	}

	return meshOf(kept), nil
}

// Section 8 groups faces into regions the intersection curve does not cross
// and settles each with a handful of rays, rather than giving every face its
// own ray as section 7 reads on its own.
func classifyBoth(
	splitA []face, idsA [][3]int, touchingA []bool, weldA *welder,
	splitB []face, idsB [][3]int, touchingB []bool, weldB *welder,
	corners []vector3.Float64,
	tolerance float64,
) (answersA, answersB []classification) {
	patchesA := patchesOf(idsA, curvePoints(weldA, corners), touchingA)
	patchesB := patchesOf(idsB, curvePoints(weldB, corners), touchingB)

	// Each solid is cast against by the other's patches, so that is the count
	// that decides whether indexing it pays.
	solidA := newSolid(splitA, tolerance, rayBudget(patchesB))
	solidB := newSolid(splitB, tolerance, rayBudget(patchesA))

	return classifyPatches(splitA, patchesA, solidB),
		classifyPatches(splitB, patchesB, solidA)
}
