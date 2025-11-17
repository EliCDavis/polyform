package marching

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/vector/vector3"
)

func marchRecurse(field sample.Vec3ToFloat, bounds geometry.AABB, cubeSize, surface float64, res map[vector3.Int]float64) {
	size := bounds.Size()
	diagonal := size.Length()

	center := bounds.Center()
	centerIndex := center.DivByConstant(cubeSize).RoundToInt()
	recentered := centerIndex.ToFloat64().Scale(cubeSize)

	fieldResult := field(recentered) - surface

	// TODO: WE THIS IS OUR BIGGEST SPEEDUP, FIGURE OUT HOW TO PRUNE HARDER
	// The closest surface is not within the bounds
	if math.Abs(fieldResult) > (diagonal/2)+(cubeSize)+center.Distance(recentered) {
		return
	}

	res[centerIndex] = fieldResult
	if size.MaxComponent() < cubeSize {
		return
	}

	halfSize := size.Scale(0.5)
	qs := halfSize.Scale(0.5)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(qs.X(), -qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), qs.Z())), halfSize), cubeSize, surface, res)
	marchRecurse(field, geometry.NewAABB(center.Add(vector3.New(-qs.X(), -qs.Y(), -qs.Z())), halfSize), cubeSize, surface, res)
}

func dedup(data *workingData, vert vector3.Float64, size float64) int {
	distritized := modeling.Vector3ToInt(vert, 4)

	if foundIndex, ok := data.vertLookup[distritized]; ok {
		return foundIndex
	}

	index := len(data.verts)
	data.vertLookup[distritized] = index
	data.verts = append(data.verts, vert.Scale(size))
	return index
}

func March(field sample.Vec3ToFloat, domain geometry.AABB, cubeSize, surface float64) modeling.Mesh {
	results := make(map[vector3.Int]float64)
	// sdfCompute := time.Now()
	marchRecurse(field, domain, cubeSize, surface, results)
	// log.Printf("Time To Compute SDFs %s", time.Since(sdfCompute))

	// marchCompute := time.Now()
	marchingWorkingData := &workingData{
		tris:       make([]int, 0),
		verts:      make([]vector3.Float64, 0),
		vertLookup: make(map[vector3.Int]int),
	}

	cubeCorners := make([]float64, 8)
	cubeCornerPositions := make([]vector3.Float64, 8)
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

		if lookupIndex == 0 || lookupIndex == 255 {
			continue
		}

		xf := float64(key.X())
		yf := float64(key.Y())
		zf := float64(key.Z())

		cubeCornerPositions[0] = vector3.New(xf, yf, zf)
		cubeCornerPositions[1] = vector3.New(xf+1, yf, zf)
		cubeCornerPositions[2] = vector3.New(xf+1, yf, zf+1)
		cubeCornerPositions[3] = vector3.New(xf, yf, zf+1)
		cubeCornerPositions[4] = vector3.New(xf, yf+1, zf)
		cubeCornerPositions[5] = vector3.New(xf+1, yf+1, zf)
		cubeCornerPositions[6] = vector3.New(xf+1, yf+1, zf+1)
		cubeCornerPositions[7] = vector3.New(xf, yf+1, zf+1)

		tris := triangulation[lookupIndex]
		for i := 0; tris[i] != -1; i += 3 {
			// Get indices of corner points A and B for each of the three edges
			// of the cube that need to be joined to form the triangle.
			a0 := cornerIndexAFromEdge[tris[i]]
			b0 := cornerIndexBFromEdge[tris[i]]

			a1 := cornerIndexAFromEdge[tris[i+1]]
			b1 := cornerIndexBFromEdge[tris[i+1]]

			a2 := cornerIndexAFromEdge[tris[i+2]]
			b2 := cornerIndexBFromEdge[tris[i+2]]

			v1 := interpolateVerts(cubeCornerPositions[a0], cubeCornerPositions[b0], cubeCorners[a0], cubeCorners[b0], 0)
			v2 := interpolateVerts(cubeCornerPositions[a1], cubeCornerPositions[b1], cubeCorners[a1], cubeCorners[b1], 0)
			v3 := interpolateVerts(cubeCornerPositions[a2], cubeCornerPositions[b2], cubeCorners[a2], cubeCorners[b2], 0)

			marchingWorkingData.tris = append(
				marchingWorkingData.tris,
				dedup(marchingWorkingData, v1, cubeSize),
				dedup(marchingWorkingData, v2, cubeSize),
				dedup(marchingWorkingData, v3, cubeSize),
			)
		}
	}

	m := modeling.NewMesh(modeling.TriangleTopology, marchingWorkingData.tris).
		SetFloat3Attribute(modeling.PositionAttribute, marchingWorkingData.verts)

	if len(marchingWorkingData.tris) == 0 {
		return m
	}

	// log.Printf("Time To March Mesh %s", time.Since(marchCompute))

	return meshops.RemoveNullFaces3D(m, modeling.PositionAttribute, 0)
}
