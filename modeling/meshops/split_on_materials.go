package meshops

import (
	"github.com/EliCDavis/polyform/modeling"
)

func SplitOnUniqueMaterials(m modeling.Mesh) []modeling.Mesh {
	// No materials to split on, so nothing to do!
	if len(m.Materials()) < 2 {
		return []modeling.Mesh{m}
	}

	// Right now we only support triangle meshes
	// (https://github.com/EliCDavis/polyform/issues/10)
	err := RequireTopology(m, modeling.TriangleTopology)
	if err != nil {
		panic(err)
	}

	type workingMesh struct {
		indices  []int
		material modeling.MeshMaterial
	}

	workingMeshes := make(map[*modeling.Material]*workingMesh)
	orderInserted := make(map[*modeling.Material]int)

	curMatIndex := 0
	trisFromOtherMats := 0

	originalMaterials := m.Materials()

	workingMeshes[originalMaterials[curMatIndex].Material] = &workingMesh{
		material: modeling.MeshMaterial{
			PrimitiveCount: 0,
			Material:       originalMaterials[curMatIndex].Material,
		},
		indices: make([]int, 0),
	}
	orderInserted[originalMaterials[curMatIndex].Material] = 0

	orinalIndices := m.Indices()
	for triStart := 0; triStart < orinalIndices.Len(); triStart += 3 {
		if originalMaterials[curMatIndex].PrimitiveCount+trisFromOtherMats <= triStart/3 {
			trisFromOtherMats += originalMaterials[curMatIndex].PrimitiveCount
			curMatIndex++
			if _, ok := workingMeshes[originalMaterials[curMatIndex].Material]; !ok {
				workingMeshes[originalMaterials[curMatIndex].Material] = &workingMesh{
					material: modeling.MeshMaterial{
						PrimitiveCount: 0,
						Material:       originalMaterials[curMatIndex].Material,
					},
					indices: make([]int, 0),
				}
				orderInserted[originalMaterials[curMatIndex].Material] = len(orderInserted)
			}
		}
		mesh := workingMeshes[originalMaterials[curMatIndex].Material]
		mesh.indices = append(
			mesh.indices,
			orinalIndices.At(triStart),
			orinalIndices.At(triStart+1),
			orinalIndices.At(triStart+2),
		)
		mesh.material.PrimitiveCount += 1
	}

	v4Data := readAllFloat4Data(m)
	v3Data := readAllFloat3Data(m)
	v2Data := readAllFloat2Data(m)
	v1Data := readAllFloat1Data(m)

	finalMeshes := make([]modeling.Mesh, len(workingMeshes))
	for mat, workingMesh := range workingMeshes {
		mesh := modeling.NewMesh(m.Topology(), workingMesh.indices).
			SetFloat4Data(v4Data).
			SetFloat3Data(v3Data).
			SetFloat2Data(v2Data).
			SetFloat1Data(v1Data).
			SetMaterial(*mat)
		finalMeshes[orderInserted[mat]] = RemovedUnreferencedVertices(mesh)
	}
	return finalMeshes
}
