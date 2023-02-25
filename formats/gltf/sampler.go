package gltf

type SamplerMagFilter int

const (
	SamplerMagFilter_NEAREST SamplerMagFilter = 9728
	SamplerMagFilter_LINEAR  SamplerMagFilter = 9729
)

type SamplerMinFilter int

const (
	SamplerMinFilter_NEAREST                SamplerMinFilter = 9728
	SamplerMinFilter_LINEAR                 SamplerMinFilter = 9729
	SamplerMinFilter_NEAREST_MIPMAP_NEAREST SamplerMinFilter = 9984
	SamplerMinFilter_LINEAR_MIPMAP_NEAREST  SamplerMinFilter = 9985
	SamplerMinFilter_NEAREST_MIPMAP_LINEAR  SamplerMinFilter = 9986
	SamplerMinFilter_LINEAR_MIPMAP_LINEAR   SamplerMinFilter = 9987
)

type SamplerWrap int

const (
	SamplerWrap_CLAMP_TO_EDGE   SamplerWrap = 33071
	SamplerWrap_MIRRORED_REPEAT SamplerWrap = 33648
	SamplerWrap_REPEAT          SamplerWrap = 10497
)

// Texture sampler properties for filtering and wrapping modes.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/sampler.schema.json
type Sampler struct {
	ChildOfRootProperty
	MagFilter SamplerMagFilter `json:"magFilter,omitempty"` // Magnification filter.
	MinFilter SamplerMinFilter `json:"minFilter,omitempty"` // Minification filter
	WrapS     SamplerWrap      `json:"wrapS,omitempty"`     // S (U) wrapping mode.  All valid values correspond to WebGL enums.
	WrapT     SamplerWrap      `json:"wrapT,omitempty"`     // T (V) wrapping mode.
}
