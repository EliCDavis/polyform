package gltf

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
)

// materialEntry tracks a unique material and its corresponding GLTF material index
type materialEntry struct {
	polyMaterial *PolyformMaterial
	index        int
}

// materialIndices handle deduplication of GLTF materials
type materialIndices []materialEntry

// findExistingMaterialID retrieves ID of the material in the tracker, if it exists.
// This is using a slice and not a map because there may be different instances of materials that have exactly
// same properties. Those need to be compared and deduplicated as well, so detailed comparison routine is used,
// instead of simple pointer equality.
func (mt materialIndices) findExistingMaterialID(mat *PolyformMaterial) (*int, bool) {
	for _, entry := range mt {
		if entry.polyMaterial.equal(mat) {
			return &entry.index, true
		}
	}
	return nil, false
}

// meshEntry tracks a unique material and its corresponding GLTF material index
type meshEntry struct {
	polyMesh      *modeling.Mesh
	materialIndex int // -1 is a valid value for absence of material
}

// meshIndices handle deduplication of GLTF meshes
type meshIndices map[meshEntry]int

// textureIndices handle deduplication of textures
type textureIndices map[*PolyformTexture]int

type writtenMeshData struct {
	attribute map[string]GltfId
	indices   *GltfId
}

type attributeIndices map[*modeling.Mesh]writtenMeshData

//=============================================================================

type instancesCachceKey struct {
	mesh int // Mesh is a combination of primitives and materials
}

type instancesCachce map[instancesCachceKey][]trs.TRS

func (ic instancesCachce) Add(mesh int, model *PolyformModel) {
	key := instancesCachceKey{mesh: mesh}
	arr := ic[key]

	if len(model.GpuInstances) == 0 {
		if model.TRS != nil {
			arr = append(arr, *model.TRS)
		} else {
			arr = append(arr, trs.Identity())
		}
		ic[key] = arr
		return
	}

	if model.TRS == nil {
		ic[key] = append(arr, model.GpuInstances...)
		return
	}

	transformedInstances := make([]trs.TRS, len(model.GpuInstances))
	for i, v := range model.GpuInstances {
		transformedInstances[i] = model.TRS.Multiply(v)
	}

	ic[key] = append(arr, transformedInstances...)
}

func (ic instancesCachce) IsInstanced(mesh int) bool {
	key := instancesCachceKey{mesh: mesh}
	return len(ic[key]) > 1
}
