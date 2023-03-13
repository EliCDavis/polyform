package voxelize

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func Vertices(mesh modeling.Mesh, attribute string, voxelSize float64) []vector3.Float64 {
	voxelSet := make(map[vector3.Int]struct{})

	// Calculate what voxel each vertex belongs to
	mesh.ScanFloat3Attribute(attribute, func(i int, v vector3.Float64) {
		voxelSet[v.DivByConstant(voxelSize).RoundToInt()] = struct{}{}
	})

	// Scale back the voxel positions back to the vertex positions
	finalValues := make([]vector3.Float64, 0, len(voxelSet))
	for voxel := range voxelSet {
		finalValues = append(finalValues, voxel.ToFloat64().Scale(voxelSize))
	}
	return finalValues
}

