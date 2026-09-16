package csg

import (
	"fmt"
	"slices"
	"strings"

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
// can. A mesh qualifies when every edge is shared by exactly two triangles of
// nonzero area, wound so the edge runs opposite ways in the two.
func CheckClosed(m modeling.Mesh) error {
	faces, err := facesOf(m)
	if err != nil {
		return err
	}
	if len(faces) == 0 {
		return fmt.Errorf("mesh has no triangles")
	}

	err = closed(faces, toleranceFor(faces, nil))
	if dropped := m.PrimitiveCount() - len(faces); err != nil && dropped > 0 {
		return fmt.Errorf("%w (%d zero-area triangles were dropped first)", err, dropped)
	}
	return err
}

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

// Directed, not unordered: section 7 reads normals, so two neighbours wound
// the same way are as broken as a hole.
func closedFrom(cornerIDs [][3]int) error {
	// Packed as (low, high, runs low to high), so sorting brings the two
	// directions of one edge together.
	edges := make([]uint64, 0, len(cornerIDs)*3)
	for _, corners := range cornerIDs {
		for k := 0; k < 3; k++ {
			from, to := corners[k], corners[(k+1)%3]
			if from == to {
				continue
			}
			runsUpward := uint64(1)
			if from > to {
				from, to, runsUpward = to, from, 0
			}
			edges = append(edges, uint64(from)<<33|uint64(to)<<1|runsUpward)
		}
	}
	slices.Sort(edges)

	dangling, crowded, flipped := 0, 0, 0
	for i := 0; i < len(edges); {
		edge := edges[i] >> 1
		upward, downward := 0, 0
		for ; i < len(edges) && edges[i]>>1 == edge; i++ {
			if edges[i]&1 == 1 {
				upward++
			} else {
				downward++
			}
		}
		switch {
		case upward+downward > 2:
			crowded++
		case upward+downward == 1:
			dangling++
		case upward != downward:
			flipped++
		}
	}

	problems := make([]string, 0, 3)
	if dangling > 0 {
		problems = append(problems, fmt.Sprintf("%d edges border a hole", dangling))
	}
	if crowded > 0 {
		problems = append(problems, fmt.Sprintf("%d edges are shared by more than two faces", crowded))
	}
	if flipped > 0 {
		problems = append(problems, fmt.Sprintf("%d edges are shared by two faces wound the same way", flipped))
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("mesh is not a closed solid: %s", strings.Join(problems, ", "))
}
