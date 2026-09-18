package csg

import (
	"fmt"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
)

// Faces meeting at a seam rarely have bit-identical corners, so positions are
// merged within the package tolerance.
type welder = geometry.PointWelder3D

func newWelder(tolerance float64) *welder {
	return geometry.NewPointWelder3D(tolerance)
}

// CheckClosed reports why a mesh cannot be a solid, or nil. Every edge must
// be shared by two triangles wound opposite ways, and the surface must face out.
func CheckClosed(m modeling.Mesh) error {
	faces, err := facesOf(m)
	if err != nil {
		return err
	}
	return validate(m, faces, weldFaces(newWelder(toleranceFor(faces, nil)), faces))
}

// Zero-area triangles facesOf dropped are named in the error: they are how a
// mesh that looks sealed turns out not to be.
func validate(m modeling.Mesh, faces []face, cornerIDs [][3]int) error {
	var err error
	if len(faces) == 0 {
		err = fmt.Errorf("mesh has no triangles")
	} else if err = edgesPairUp(cornerIDs); err == nil && signedVolume(faces) < 0 {
		err = fmt.Errorf("mesh is wound inside out")
	}
	if dropped := m.PrimitiveCount() - len(faces); err != nil && dropped > 0 {
		return fmt.Errorf("%w (%d zero-area triangles were dropped first)", err, dropped)
	}
	return err
}

// Divergence theorem: positive when the faces wind outward, which is what
// lets section 7 read a normal as pointing away from the inside.
func signedVolume(faces []face) float64 {
	total := 0.
	for _, f := range faces {
		total += f.verts[0].Dot(f.verts[1].Cross(f.verts[2]))
	}
	return total / 6
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

// Every edge must be used by exactly two faces running opposite ways.
// Neighbours wound the same way are as broken as a hole.
func edgesPairUp(cornerIDs [][3]int) error {
	// One integer per directed edge: low id, high id, and a direction bit.
	// Sorting lands both directions of an edge side by side.
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

	// Each run of equal ids is one edge; a sound one has one of each direction.
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
