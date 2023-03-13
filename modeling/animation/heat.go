package animation

import (
	"container/heap"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type HeatSource struct {
	position  vector3.Float64
	intensity float64
}

var cubeIndices = [][]int{
	{-1, -1, -1},
	{-1, -1, 0},
	{-1, -1, 1},
	{-1, 0, -1},
	{-1, 0, 0},
	{-1, 0, 1},
	{-1, 1, -1},
	{-1, 1, 0},
	{-1, 1, 1},

	{0, -1, -1},
	{0, -1, 0},
	{0, -1, 1},
	{0, 0, -1},
	{0, 0, 0},
	{0, 0, 1},
	{0, 1, -1},
	{0, 1, 0},
	{0, 1, 1},

	{1, -1, -1},
	{1, -1, 0},
	{1, -1, 1},
	{1, 0, -1},
	{1, 0, 0},
	{1, 0, 1},
	{1, 1, -1},
	{1, 1, 0},
	{1, 1, 1},
}

func avgValSurroundingPos(pos vector3.Int, workingData map[vector3.Int]float64) float64 {
	// cellsUsed := 0
	totalHeat := 0.
	for _, cubeOffset := range cubeIndices {
		newPos := pos.Add(vector3.New(cubeOffset[0], cubeOffset[1], cubeOffset[2]))
		if val, ok := workingData[newPos]; ok {
			// cellsUsed++
			totalHeat += val
		}
	}
	return totalHeat / 27
}

func diffuseHeat(voxelization []vector3.Float64, voxelSize float64, heatSource HeatSource, iterations int) map[vector3.Int]float64 {
	workingData := make(map[vector3.Int]float64)

	// Initialize Working data all to 0
	for _, v := range voxelization {
		workingData[v.DivByConstant(voxelSize).RoundToInt()] = 0
	}

	heatSourcePos := heatSource.position.DivByConstant(voxelSize).RoundToInt()

	// Propagate heat through voxel data
	for i := 0; i < iterations; i++ {

		// Keep the heating element at it's peak
		workingData[heatSourcePos] = heatSource.intensity

		// Set each voxel value to be the average of the surrounding voxels
		for voxel := range workingData {
			workingData[voxel] = avgValSurroundingPos(voxel, workingData)
		}
	}

	return workingData
}

func WeightMeshWithHeatDiffusion(mesh modeling.Mesh, skeleton Skeleton, voxilization []vector3.Float64, voxelSize float64, iterations int) modeling.Mesh {
	const maxPointsToConsider = 4

	// voxilization := voxelize.Surface(mesh, modeling.PositionAttribute, voxelSize)

	jointHeat := make([]map[vector3.Int]float64, skeleton.JointCount())
	for i := 0; i < skeleton.JointCount(); i++ {
		heatSource := HeatSource{
			position:  skeleton.WorldPosition(i),
			intensity: skeleton.Heat(i),
		}
		jointHeat[i] = diffuseHeat(voxilization, voxelSize, heatSource, iterations)
	}

	weightData := make([]vector4.Float64, mesh.AttributeLength())
	jointData := make([]vector4.Float64, mesh.AttributeLength())

	return mesh.
		ScanFloat3Attribute(modeling.PositionAttribute, func(vertexIndex int, v vector3.Float64) {
			cell := v.DivByConstant(voxelSize).RoundToInt()
			queue := make(maxJointValPriorityQueue, 0)

			for jointIndex, heatLUT := range jointHeat {
				heap.Push(&queue, jointValItem{
					val:   heatLUT[cell],
					joint: jointIndex,
				})
			}

			size := maxPointsToConsider
			if queue.Len() < maxPointsToConsider {
				size = queue.Len()
			}

			joints := make([]int, size)
			jointVals := make([]float64, size)
			totalVal := 0.
			for i := 0; i < size; i++ {
				item := heap.Pop(&queue).(jointValItem)
				joints[i] = item.joint
				jointVals[i] = item.val
				totalVal += item.val
			}

			if totalVal == 0 {
				jointData[vertexIndex] = vector4.New(float64(joints[0]), 0., 0., 0.)
				weightData[vertexIndex] = vector4.New(1., 0., 0., 0.)
				return
			}

			jointData[vertexIndex] = vector4.New(float64(joints[0]), float64(joints[1]), float64(joints[2]), float64(joints[3]))
			weightData[vertexIndex] = vector4.New(jointVals[0], jointVals[1], jointVals[2], jointVals[3]).DivByConstant(totalVal)
		}).
		SetFloat4Attribute(modeling.WeightAttribute, weightData).
		SetFloat4Attribute(modeling.JointAttribute, jointData)
}
