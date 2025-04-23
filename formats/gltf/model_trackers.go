package gltf

import (
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
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

// materialEntry tracks a unique material and its corresponding GLTF material index
type meshEntry struct {
	polyMesh      *modeling.Mesh
	materialIndex int // -1 is a valid value for absence of material
}

// materialIndices handle deduplication of GLTF materials
type meshIndices map[meshEntry]int

// textureIndices handle deduplication of textures
type textureIndices map[*PolyformTexture]int

type writtenMeshData struct {
	attribute map[string]GltfId
	indices   *GltfId
}

type attributeIndices map[*modeling.Mesh]writtenMeshData

// modelInstance represents an instance of a model with specific transformation
type modelInstance struct {
	meshIndex int
	name      string
	trs       *trs.TRS // Holds the transformation data for this instance
}

// modelInstanceGroup contains instances of the same mesh that should be rendered together
type modelInstanceGroup struct {
	meshIndex  int
	instances  []modelInstance
	skeleton   *animation.Skeleton
	animations []animation.Sequence
}

func (g modelInstanceGroup) isAnimated() bool {
	return g.skeleton != nil || len(g.animations) > 0
}

// instanceTracker handles tracking and organizing model instances
type instanceTracker struct {
	groups []modelInstanceGroup
}

// findOrCreateGroup finds an existing group with matching mesh and material indices or creates a new one
func (it *instanceTracker) findOrCreateGroup(meshIndex int, skeleton *animation.Skeleton, animations []animation.Sequence) *modelInstanceGroup {
	// For models with skeletons or animations, we need to create a unique group
	// because these can't be instanced with other models
	hasAnimated := skeleton != nil || len(animations) > 0

	if hasAnimated {
		// Create a new unique group for this animated model
		it.groups = append(it.groups, modelInstanceGroup{
			meshIndex:  meshIndex,
			instances:  make([]modelInstance, 0),
			skeleton:   skeleton,
			animations: animations,
		})
		return &it.groups[len(it.groups)-1]
	}

	// For standard models without animations, look for an existing group with the same mesh
	for i := range it.groups {
		g := &it.groups[i]
		if g.meshIndex == meshIndex && !g.isAnimated() {
			return g
		}
	}

	// No existing group found, create a new one
	it.groups = append(it.groups, modelInstanceGroup{
		meshIndex: meshIndex,
		instances: make([]modelInstance, 0),
	})

	return &it.groups[len(it.groups)-1]
}

// add adds a new instance to the appropriate group
func (it *instanceTracker) add(meshIndex int, instance modelInstance, skeleton *animation.Skeleton, animations []animation.Sequence) {
	group := it.findOrCreateGroup(meshIndex, skeleton, animations)
	group.instances = append(group.instances, instance)
}
