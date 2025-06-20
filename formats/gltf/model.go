package gltf

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
)

type PolyformScene struct {
	Models []PolyformModel
	Lights []KHR_LightsPunctual

	// UseGpuInstancing indicates that the EXT_mesh_gpu_instancing extension should be used
	// when appropriate for mesh instances. This must be set if GPU instances are defined on any of the models in the scene.
	// If not set while GPU instances are defined - the scene serialisation will fail.
	// This is a global setting for the scene and cannot be set on a per-model basis.
	//
	// The model deduplication applies regardless, and models that have exactly the same mesh pointer reference,
	// and material value will be collapsed into a single list.
	// If this flag is set, this single list will be converted into GPU instances as well.
	UseGpuInstancing bool
}

// PolyformModel is a utility structure for reading/writing to GLTF format within
// polyform, and not an actual concept found within the GLTF format.
type PolyformModel struct {
	Name     string
	Mesh     *modeling.Mesh
	Material *PolyformMaterial

	// TRS contains the transformation (translation, rotation, scale) for this model
	// This is optional and it will be used if the models are deduplicated and collapsed into a list of instances.
	TRS *trs.TRS

	// Utilizes the EXT_mesh_gpu_instancing extension to duplicate the model
	// without increasing the mesh data footprint on the GPU.
	// This is a list of transformations where this model should be repeated.
	// This can only used if the UseGpuInstancing flag is set on the scene.
	// If flag is not set, populating this list will cause the scene writing to fail.
	GpuInstances []trs.TRS

	// Animation is a list of animations that are applied to this model.
	// Models with animations defined will never be deduplicated into a single list.
	// However, these models can utilize GpuInstances and in that case the same animation will be applied to all of them.
	// Limitations on using GpuInstances still apply.
	Skeleton   *animation.Skeleton
	Animations []animation.Sequence
}

type PolyformMaterial struct {
	Name                 string
	Extras               map[string]any
	AlphaMode            *MaterialAlphaMode
	AlphaCutoff          *float64
	PbrMetallicRoughness *PolyformPbrMetallicRoughness
	Extensions           []MaterialExtension
	NormalTexture        *PolyformNormal
	OcclusionTexture     *PolyformOcclusion
	EmissiveTexture      *PolyformTexture
	EmissiveFactor       color.Color
}

type PolyformPbrMetallicRoughness struct {
	BaseColorFactor          color.Color
	BaseColorTexture         *PolyformTexture
	MetallicFactor           *float64
	RoughnessFactor          *float64
	MetallicRoughnessTexture *PolyformTexture
}

type PolyformNormal struct {
	*PolyformTexture
	Scale *float64
}

type PolyformOcclusion struct {
	*PolyformTexture
	Strength *float64
}

type PolyformTexture struct {
	URI        string
	Image      image.Image
	Sampler    *Sampler
	Extensions []TextureExtension
}

func (pt *PolyformTexture) canAddToGLTF() bool {
	if pt == nil {
		return false
	}

	return pt.URI != "" || pt.Image != nil
}

func (pm *PolyformTexture) prepareExtensions(w *Writer) (map[string]any, map[string]any) {
	var texInfoExt map[string]any
	var texExt map[string]any

	for _, ext := range pm.Extensions {
		id := ext.ExtensionID()
		if ext.IsInfo() {
			if texInfoExt == nil {
				texInfoExt = make(map[string]any)
			}

			texInfoExt[id] = ext.ToTextureExtensionData(w)
		} else {
			if texExt == nil {
				texExt = make(map[string]any)
			}

			texExt[id] = ext.ToTextureExtensionData(w)
		}

		w.extensionsUsed[id] = true
		if ext.IsRequired() {
			w.extensionsRequired[id] = true
		}
	}

	return texExt, texInfoExt
}

func (pm *PolyformMaterial) equal(other *PolyformMaterial) bool {
	if pm == other {
		return true
	}

	if pm == nil || other == nil {
		return false
	}

	if pm.Name != other.Name {
		return false
	}
	if !pm.PbrMetallicRoughness.equal(other.PbrMetallicRoughness) {
		return false
	}
	if !pm.EmissiveTexture.equal(other.EmissiveTexture) {
		return false
	}
	if !colorsEqual(pm.EmissiveFactor, other.EmissiveFactor) {
		return false
	}

	if (pm.AlphaMode == nil) != (other.AlphaMode == nil) {
		return false
	} else if pm.AlphaMode != nil && other.AlphaMode != nil && *pm.AlphaMode != *other.AlphaMode {
		return false
	}

	if !float64PtrsEqual(pm.AlphaCutoff, other.AlphaCutoff) {
		return false
	}
	if len(pm.Extensions) != len(other.Extensions) {
		return false
	}
	for i, ext := range pm.Extensions {
		if ext != other.Extensions[i] {
			return false
		}
	}
	return true
}

func (pt *PolyformTexture) equal(other *PolyformTexture) bool {
	if pt == other {
		return true
	}

	if pt == nil || other == nil {
		return false
	}

	if pt.URI != other.URI {
		return false
	}

	if pt.Sampler == other.Sampler {
		return true
	} else if pt.Sampler == nil || other.Sampler == nil {
		return false
	}

	if pt.Sampler.MagFilter != other.Sampler.MagFilter ||
		pt.Sampler.MinFilter != other.Sampler.MinFilter ||
		pt.Sampler.WrapS != other.Sampler.WrapS ||
		pt.Sampler.WrapT != other.Sampler.WrapT {
		return false
	}

	return true
}

func (pt *PolyformNormal) equal(other *PolyformNormal) bool {
	if pt == other {
		return true
	}

	if pt == nil || other == nil {
		return false
	}

	if !pt.PolyformTexture.equal(other.PolyformTexture) {
		return false
	}
	return float64PtrsEqual(pt.Scale, other.Scale)
}

func (pmr *PolyformPbrMetallicRoughness) equal(other *PolyformPbrMetallicRoughness) bool {
	if pmr == other {
		return true
	}

	if pmr == nil || other == nil {
		return false
	}

	if !float64PtrsEqual(pmr.MetallicFactor, other.MetallicFactor) ||
		!float64PtrsEqual(pmr.RoughnessFactor, other.RoughnessFactor) ||
		!colorsEqual(pmr.BaseColorFactor, other.BaseColorFactor) {
		return false
	}

	if !pmr.BaseColorTexture.equal(other.BaseColorTexture) ||
		!pmr.MetallicRoughnessTexture.equal(other.MetallicRoughnessTexture) {
		return false
	}

	return true
}

// Helper functions for comparing nullable values
func float64PtrsEqual(a, b *float64) bool {
	if a == b {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func colorsEqual(a, b color.Color) bool {
	if a == b {
		return true
	} else if a == nil || b == nil {
		return false
	}
	// Since color.Color is an interface, we can only check for basic RGBA equality
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
