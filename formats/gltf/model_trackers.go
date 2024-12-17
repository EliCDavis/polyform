package gltf

import "github.com/EliCDavis/polyform/modeling"

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

// materialEntry tracks a unique material and its corresponding GLTF material index
type meshEntry struct {
	polyMesh      *modeling.Mesh
	materialIndex int // -1 is a valid value for absence of material
}

// materialIndices handle deduplication of GLTF materials
type meshIndices map[meshEntry]int

// textureIndices handle deduplication of textures
type textureIndices map[*PolyformTexture]int
