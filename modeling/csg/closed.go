package csg

import (
	"fmt"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
)

// Faces arrive unwelded and two of them meeting at a seam rarely carry bit
// identical corners, so positions are merged within the same tolerance the
// rest of the package works to.
type welder = geometry.PointWelder3D

func newWelder(tolerance float64) *welder {
	return geometry.NewPointWelder3D(tolerance)
}

// CheckClosed reports why a mesh cannot be used as a solid, or nil when it
// can. A mesh qualifies when every edge is shared by exactly two triangles,
// which is what makes "inside" mean anything.
func CheckClosed(m modeling.Mesh) error {
	faces, err := facesOf(m)
	if err != nil {
		return err
	}
	if len(faces) == 0 {
		return fmt.Errorf("mesh has no triangles")
	}
	return closed(faces, toleranceFor(faces, nil))
}

// Section 7 decides inside from outside by what a ray leaving a face hits
// first. Through a hole that answer is whatever happens to be behind it, so
// an open mesh does not fail loudly further down - it quietly returns
// nonsense. Hence the check up front.
// The ids outlive the check. Section 3 holds one vertex array with polygons
// pointing into it, so which faces meet is known without asking where their
// corners are; everything downstream that needs adjacency reads these rather
// than paying to rediscover them by position.
func weldFaces(weld *welder, faces []face) [][3]int {
	ids := make([][3]int, len(faces))
	for i, f := range faces {
		ids[i] = [3]int{
			weld.Index(f.verts[0]),
			weld.Index(f.verts[1]),
			weld.Index(f.verts[2]),
		}
	}
	return ids
}

func closed(faces []face, tolerance float64) error {
	return closedFrom(weldFaces(newWelder(tolerance), faces))
}

func closedFrom(ids [][3]int) error {
	edges := make(map[[2]int]int, len(ids)*3)

	for _, corners := range ids {
		for k := 0; k < 3; k++ {
			a, b := corners[k], corners[(k+1)%3]
			if a == b {
				continue
			}
			if a > b {
				a, b = b, a
			}
			edges[[2]int{a, b}]++
		}
	}

	dangling, crowded := 0, 0
	for _, shared := range edges {
		switch {
		case shared == 1:
			dangling++
		case shared > 2:
			crowded++
		}
	}

	switch {
	case dangling > 0 && crowded > 0:
		return fmt.Errorf(
			"mesh is not a closed solid: %d edges border a hole and %d are shared by more than two faces",
			dangling, crowded)
	case dangling > 0:
		return fmt.Errorf("mesh is not a closed solid: %d edges border a hole", dangling)
	case crowded > 0:
		return fmt.Errorf(
			"mesh is not a closed solid: %d edges are shared by more than two faces",
			crowded)
	}
	return nil
}
