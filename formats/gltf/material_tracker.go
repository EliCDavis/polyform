package gltf

// materialEntry tracks a unique material and its corresponding GLTF material index
type materialEntry struct {
	polyMaterial *PolyformMaterial
	index        int
}

// materialTracker handles deduplication of GLTF materials
type materialTracker struct {
	entries []materialEntry
}

// findExistingMaterialID retrieves ID of the material in the tracker, if it exists
func (mt *materialTracker) findExistingMaterialID(mat *PolyformMaterial) (*int, bool) {
	for _, entry := range mt.entries {
		if entry.polyMaterial.equal(mat) {
			return &entry.index, true
		}
	}
	return nil, false
}
