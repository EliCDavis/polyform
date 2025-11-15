package marching

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func marchRecurse(field sample.Vec3ToFloat, bounds geometry.AABB, cubeSize float64, res map[vector3.Int]float64) {
	center := bounds.Center()
	size := bounds.Size()

	// The closest surface is not within the bounds
	fieldResult := field(center)
	if math.Abs(fieldResult) > (size.MaxComponent()/2)+(cubeSize*2) {
		return
	}

	if size.MaxComponent() > cubeSize {
		halfSize := size.Scale(0.5)
		qs := halfSize.Scale(0.5)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), -qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), -qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), -qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), qs.Z())), halfSize), cubeSize, res)
		marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), -qs.Z())), halfSize), cubeSize, res)
		return
	}

	res[center.DivByConstant(cubeSize).FloorToInt()] = fieldResult
}

func dedup(data *workingData, vert vector3.Float64, size float64) int {
	distritized := vert.ToInt()

	if foundIndex, ok := data.vertLookup[distritized]; ok {
		return foundIndex
	}

	index := len(data.verts)
	data.vertLookup[distritized] = index
	data.verts = append(data.verts, vert.Scale(size))
	return index
}

func March(field sample.Vec3ToFloat, domain geometry.AABB, cubeSize float64) modeling.Mesh {
	results := make(map[vector3.Int]float64)
	marchRecurse(field, domain, cubeSize, results)

	marchingWorkingData := &workingData{
		tris:       make([]int, 0),
		verts:      make([]vector3.Float64, 0),
		vertLookup: make(map[vector3.Int]int),
	}

	// tris := make([]int, 0)
	// verts := make([]vector3.Float64, 0)

	cubeCorners := make([]float64, 8)
	for key, nnn := range results {
		cubeCorners[0] = nnn

		var ok bool
		cubeCorners[1], ok = results[key.Add(vector3.New(1, 0, 0))]
		if !ok {
			continue
		}
		cubeCorners[2], ok = results[key.Add(vector3.New(1, 0, 1))]
		if !ok {
			continue
		}
		cubeCorners[3], ok = results[key.Add(vector3.New(0, 0, 1))]
		if !ok {
			continue
		}
		cubeCorners[4], ok = results[key.Add(vector3.New(0, 1, 0))]
		if !ok {
			continue
		}
		cubeCorners[5], ok = results[key.Add(vector3.New(1, 1, 0))]
		if !ok {
			continue
		}
		cubeCorners[6], ok = results[key.Add(vector3.New(1, 1, 1))]
		if !ok {
			continue
		}
		cubeCorners[7], ok = results[key.Add(vector3.New(0, 1, 1))]
		if !ok {
			continue
		}

		lookupIndex := 0
		if cubeCorners[0] < 0 {
			lookupIndex |= 1
		}
		if cubeCorners[1] < 0 {
			lookupIndex |= 2
		}
		if cubeCorners[2] < 0 {
			lookupIndex |= 4
		}
		if cubeCorners[3] < 0 {
			lookupIndex |= 8
		}
		if cubeCorners[4] < 0 {
			lookupIndex |= 16
		}
		if cubeCorners[5] < 0 {
			lookupIndex |= 32
		}
		if cubeCorners[6] < 0 {
			lookupIndex |= 64
		}
		if cubeCorners[7] < 0 {
			lookupIndex |= 128
		}

		xf := float64(key.X())
		yf := float64(key.Y())
		zf := float64(key.Z())

		cubeCornerPositions := []vector3.Float64{
			vector3.New(xf, yf, zf),
			vector3.New(xf+1, yf, zf),
			vector3.New(xf+1, yf, zf+1),
			vector3.New(xf, yf, zf+1),
			vector3.New(xf, yf+1, zf),
			vector3.New(xf+1, yf+1, zf),
			vector3.New(xf+1, yf+1, zf+1),
			vector3.New(xf, yf+1, zf+1),
		}

		for i := 0; triangulation[lookupIndex][i] != -1; i += 3 {
			// Get indices of corner points A and B for each of the three edges
			// of the cube that need to be joined to form the triangle.
			a0 := cornerIndexAFromEdge[triangulation[lookupIndex][i]]
			b0 := cornerIndexBFromEdge[triangulation[lookupIndex][i]]

			a1 := cornerIndexAFromEdge[triangulation[lookupIndex][i+1]]
			b1 := cornerIndexBFromEdge[triangulation[lookupIndex][i+1]]

			a2 := cornerIndexAFromEdge[triangulation[lookupIndex][i+2]]
			b2 := cornerIndexBFromEdge[triangulation[lookupIndex][i+2]]

			v1 := interpolateVerts(cubeCornerPositions[a0], cubeCornerPositions[b0], cubeCorners[a0], cubeCorners[b0], 0)
			v2 := interpolateVerts(cubeCornerPositions[a1], cubeCornerPositions[b1], cubeCorners[a1], cubeCorners[b1], 0)
			v3 := interpolateVerts(cubeCornerPositions[a2], cubeCornerPositions[b2], cubeCorners[a2], cubeCorners[b2], 0)

			// verts = append(
			// 	verts,
			// 	v1.Scale(cubeSize),
			// 	v2.Scale(cubeSize),
			// 	v3.Scale(cubeSize),
			// )

			// tris = append(tris, len(tris), len(tris)+1, len(tris)+2)

			marchingWorkingData.tris = append(
				marchingWorkingData.tris,
				dedup(marchingWorkingData, v1, cubeSize),
				dedup(marchingWorkingData, v2, cubeSize),
				dedup(marchingWorkingData, v3, cubeSize),
			)
		}
	}

	return modeling.NewMesh(modeling.TriangleTopology, marchingWorkingData.tris).
		SetFloat3Attribute(modeling.PositionAttribute, marchingWorkingData.verts)
}
