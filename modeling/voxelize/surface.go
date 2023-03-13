package voxelize

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func Surface(mesh modeling.Mesh, attribute string, voxelSize float64) []vector3.Float64 {
	type tri struct{ a, b, c vector3.Float64 }

	// Initialize workload with all triangles of the mesh
	work := make([]tri, mesh.PrimitiveCount())
	for i := 0; i < mesh.PrimitiveCount(); i++ {
		meshTri := mesh.Tri(i)
		work[i] = tri{
			a: meshTri.P1Vec3Attr(attribute),
			b: meshTri.P2Vec3Attr(attribute),
			c: meshTri.P3Vec3Attr(attribute),
		}
	}

	voxelSet := make(map[vector3.Int]struct{})
	var curTri tri
	for len(work) > 0 {
		curTri, work = work[len(work)-1], work[:len(work)-1]

		a := curTri.a.DivByConstant(voxelSize).RoundToInt()
		b := curTri.b.DivByConstant(voxelSize).RoundToInt()
		c := curTri.c.DivByConstant(voxelSize).RoundToInt()

		// The triangle in question spans multiple voxels. In order to
		// guarantee we get all values in between these voxels, we subdivide
		// The triangle and try again.
		abDist := a.Distance(b)
		acDist := a.Distance(c)
		bcDist := b.Distance(c)
		longest := math.Max(abDist, math.Max(acDist, bcDist))

		// If the triangle spans too great a distance, divide it into 4 parts.
		if longest > 1.7 { // Greater than sqrt(3); dist (0,0,0  1,1,1)
			abMid := curTri.a.Midpoint(curTri.b)
			acMid := curTri.a.Midpoint(curTri.c)
			bcMid := curTri.b.Midpoint(curTri.c)
			work = append(
				work,
				tri{curTri.a, abMid, acMid},
				tri{curTri.b, abMid, bcMid},
				tri{curTri.c, bcMid, acMid},
				tri{abMid, bcMid, acMid},
			)
			continue
		}

		// All points on the triangle are neighboring voxels! We're done!
		voxelSet[a] = struct{}{}
		voxelSet[b] = struct{}{}
		voxelSet[c] = struct{}{}
	}

	// Scale back the voxel positions back to the vertex positions
	finalValues := make([]vector3.Float64, 0, len(voxelSet))
	for voxel := range voxelSet {
		finalValues = append(finalValues, voxel.ToFloat64().Scale(voxelSize))
	}
	return finalValues
}
