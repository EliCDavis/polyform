package gltf

import (
	"image/color"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
)

type PolyformScene struct {
	Models []PolyformModel
	Lights []KHR_LightsPunctual
}

// PolyformModel is a utility structure for reading/writing to GLTF format within
// polyform, and not an actual concept found within the GLTF format.
type PolyformModel struct {
	Name     string
	Mesh     modeling.Mesh
	Material *PolyformMaterial

	Skeleton   *animation.Skeleton
	Animations []animation.Sequence
}

type PolyformMaterial struct {
	Name                 string
	Extras               map[string]any
	AlphaMode            MaterialAlphaMode
	PbrMetallicRoughness *PolyformPbrMetallicRoughness
	Extensions           []MaterialExtension
	NormalTexture        *PolyformNormal
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
	PolyformTexture
	Scale *float64
}

type PolyformTexture struct {
	URI     string
	Sampler *Sampler
}
