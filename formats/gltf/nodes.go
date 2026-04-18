package gltf

import (
	"image"
	"image/color"
	"io"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ManifestNode]](factory)
	refutil.RegisterType[nodes.Struct[MaterialAnisotropyExtensionNode]](factory)
	refutil.RegisterType[nodes.Struct[MaterialClearcoatExtensionNode]](factory)
	refutil.RegisterType[nodes.Struct[MaterialNode]](factory)
	refutil.RegisterType[nodes.Struct[MaterialTransmissionExtensionNode]](factory)
	refutil.RegisterType[nodes.Struct[MaterialVolumeExtensionNode]](factory)
	refutil.RegisterType[nodes.Struct[ModelNode]](factory)
	refutil.RegisterType[nodes.Struct[TextureReferenceNode]](factory)
	refutil.RegisterType[nodes.Struct[TextureNode]](factory)
	refutil.RegisterType[nodes.Struct[NormalTextureNode]](factory)
	refutil.RegisterType[nodes.Struct[SamplerNode]](factory)

	refutil.RegisterType[nodes.Struct[AnimationNode]](factory)
	refutil.RegisterType[nodes.Struct[TranslationAnimationChannelNode]](factory)
	refutil.RegisterType[nodes.Struct[RotationAnimationChannelNode]](factory)
	refutil.RegisterType[nodes.Struct[ScaleAnimationChannelNode]](factory)

	refutil.RegisterType[SamplerWrapNode](factory)
	refutil.RegisterType[nodes.Struct[SamplerMinFilterNode]](factory)
	refutil.RegisterType[nodes.Struct[SamplerMagFilterNode]](factory)

	generator.RegisterTypes(factory)
}

type Artifact struct {
	Scene   PolyformScene
	Options WriterOptions
}

func (Artifact) Mime() string {
	return "model/gltf-binary"
}

func (ga Artifact) Write(w io.Writer) error {
	return WriteBinary(ga.Scene, w, nil)
}

type ManifestNode struct {
	Models     []nodes.Output[*PolyformModel]
	Animations []nodes.Output[PolyformAnimation]
}

func (gad ManifestNode) Out(out *nodes.StructOutput[manifest.Manifest]) {
	models := make([]*PolyformModel, 0, len(gad.Models))

	for _, m := range gad.Models {
		if m == nil {
			continue
		}
		value := nodes.GetOutputValue(out, m)

		// // TechDebt: Skip nodes without meshes as at the moment it'll cause stuff
		// // to error out
		// if value.Mesh == nil {
		// 	continue
		// }

		models = append(models, value)
	}

	entry := manifest.Entry{
		Artifact: &Artifact{
			Scene: PolyformScene{
				Models:     models,
				Animations: nodes.GetOutputValues(out, gad.Animations),
			},
			Options: WriterOptions{
				GpuInstancingStrategy: WriterInstancingStrategy_Default,
			},
		},
	}

	out.Set(manifest.SingleEntryManifest("model.glb", entry))
}

type ModelNode struct {
	Name     nodes.Output[string]
	Mesh     nodes.Output[modeling.Mesh]
	Material nodes.Output[PolyformMaterial]
	Children []nodes.Output[*PolyformModel]

	Translation nodes.Output[vector3.Float64]
	Rotation    nodes.Output[quaternion.Quaternion]
	Scale       nodes.Output[vector3.Float64]

	GpuInstances nodes.Output[[]trs.TRS]
}

func (gmnd ModelNode) Out(out *nodes.StructOutput[*PolyformModel]) {
	transform := trs.New(
		nodes.TryGetOutputValue(out, gmnd.Translation, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, gmnd.Rotation, quaternion.Identity()),
		nodes.TryGetOutputValue(out, gmnd.Scale, vector3.One[float64]()),
	)

	out.Set(&PolyformModel{
		Name:         nodes.TryGetOutputValue(out, gmnd.Name, "Mesh"),
		GpuInstances: nodes.TryGetOutputValue(out, gmnd.GpuInstances, nil),
		Material:     nodes.TryGetOutputReference(out, gmnd.Material, nil),
		Mesh:         nodes.TryGetOutputReference(out, gmnd.Mesh, nil),
		TRS:          &transform,
		Children:     nodes.GetOutputValues(out, gmnd.Children),
	})
}

type TextureReferenceNode struct {
	URI     nodes.Output[string]
	Sampler nodes.Output[Sampler]
}

func (tnd TextureReferenceNode) Out(out *nodes.StructOutput[PolyformTexture]) {
	out.Set(PolyformTexture{
		Sampler: nodes.TryGetOutputReference(out, tnd.Sampler, nil),
		URI:     nodes.TryGetOutputValue(out, tnd.URI, ""),
	})
}

func (gmnd TextureReferenceNode) Description() string {
	return "An object that combines an image and its sampler"
}

type TextureNode struct {
	Image   nodes.Output[image.Image]
	Sampler nodes.Output[Sampler]
}

func (tnd TextureNode) Out(out *nodes.StructOutput[PolyformTexture]) {
	out.Set(PolyformTexture{
		Sampler: nodes.TryGetOutputReference(out, tnd.Sampler, nil),
		Image:   nodes.TryGetOutputValue(out, tnd.Image, nil),
	})
}

func (gmnd TextureNode) Description() string {
	return "An object that combines an image and its sampler"
}

type NormalTextureNode struct {
	Texture nodes.Output[PolyformTexture]
	Scale   nodes.Output[float64]
}

func (tnd NormalTextureNode) Out(out *nodes.StructOutput[PolyformNormal]) {
	out.Set(PolyformNormal{
		Scale:           nodes.TryGetOutputReference(out, tnd.Scale, nil),
		PolyformTexture: nodes.TryGetOutputReference(out, tnd.Texture, nil),
	})
}

type SamplerNode struct {
	MagFilter nodes.Output[SamplerMagFilter] `description:"Magnification filter"`
	MinFilter nodes.Output[SamplerMinFilter] `description:"Minification filter"`
	WrapS     nodes.Output[SamplerWrap]      `description:"S (U) wrapping mode"`
	WrapT     nodes.Output[SamplerWrap]      `description:"T (V) wrapping mode"`
}

func (tnd SamplerNode) Out(out *nodes.StructOutput[Sampler]) {
	out.Set(Sampler{
		MagFilter: nodes.TryGetOutputValue(out, tnd.MagFilter, SamplerMagFilter_NEAREST),
		MinFilter: nodes.TryGetOutputValue(out, tnd.MinFilter, SamplerMinFilter_NEAREST),
		WrapS:     nodes.TryGetOutputValue(out, tnd.WrapS, SamplerWrap_REPEAT),
		WrapT:     nodes.TryGetOutputValue(out, tnd.WrapT, SamplerWrap_REPEAT),
	})
}

func (SamplerNode) Description() string {
	return "Texture sampler properties for filtering and wrapping modes"
}

type SamplerWrapNode struct {
}

func (SamplerWrapNode) Name() string {
	return "Sampler Wrap"
}

func (SamplerWrapNode) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *SamplerWrapNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Clamp To Edge": nodes.ConstOutput[SamplerWrap]{
			Ref:             p,
			Val:             SamplerWrap_CLAMP_TO_EDGE,
			PortName:        "Clamp To Edge",
			PortDescription: "The last pixel of the texture stretches to the edge of the mesh.",
		},

		"Mirrored Repeat": nodes.ConstOutput[SamplerWrap]{
			Ref:             p,
			Val:             SamplerWrap_MIRRORED_REPEAT,
			PortName:        "Mirrored Repeat",
			PortDescription: "The texture will repeats to infinity, mirroring on each repeat",
		},

		"Repeat": nodes.ConstOutput[SamplerWrap]{
			Ref:             p,
			Val:             SamplerWrap_REPEAT,
			PortName:        "Repeat",
			PortDescription: "The texture will simply repeat to infinity.",
		},
	}
}

type SamplerMagFilterNode struct {
}

func (SamplerMagFilterNode) Name() string {
	return "Sampler Mag Filter"
}

func (SamplerMagFilterNode) Description() string {
	return "These define the texture magnification function to be used when the pixel being textured maps to an area less than or equal to one texture element (texel)"
}

func (SamplerMagFilterNode) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *SamplerMagFilterNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Nearest": nodes.ConstOutput[SamplerMagFilter]{
			Ref:             p,
			Val:             SamplerMagFilter_NEAREST,
			PortName:        "Nearest",
			PortDescription: "The value of the texture element that is nearest (in Manhattan distance) to the specified texture coordinates",
		},

		"Linear": nodes.ConstOutput[SamplerMagFilter]{
			Ref:             p,
			Val:             SamplerMagFilter_LINEAR,
			PortName:        "Linear",
			PortDescription: "The weighted average of the four texture elements that are closest to the specified texture coordinates, and can include items wrapped or repeated from other parts of a texture, depending on the values of wrapS and wrapT, and on the exact mapping.",
		},
	}
}

// Descriptions pulled from Three.js
// http://threejs.org/docs/#api/en/constants/Textures

type SamplerMinFilterNode struct {
}

func (SamplerMinFilterNode) Name() string {
	return "Sampler Min Filter"
}

func (SamplerMinFilterNode) Description() string {
	return "These define the texture minifying function that is used whenever the pixel being textured maps to an area greater than one texture element (texel)."
}

func (SamplerMinFilterNode) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *SamplerMinFilterNode) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Nearest": nodes.ConstOutput[SamplerMinFilter]{
			Ref:             p,
			Val:             SamplerMinFilter_NEAREST,
			PortName:        "Nearest",
			PortDescription: "The value of the texture element that is nearest (in Manhattan distance) to the specified texture coordinates",
		},

		"Linear": nodes.ConstOutput[SamplerMinFilter]{
			Ref:             p,
			Val:             SamplerMinFilter_LINEAR,
			PortName:        "Linear",
			PortDescription: "The weighted average of the four texture elements that are closest to the specified texture coordinates, and can include items wrapped or repeated from other parts of a texture, depending on the values of wrapS and wrapT, and on the exact mapping.",
		},

		"Nearest Mipmap Nearest": nodes.ConstOutput[SamplerMinFilter]{
			Ref:             p,
			Val:             SamplerMinFilter_NEAREST_MIPMAP_NEAREST,
			PortName:        "Nearest Mipmap Nearest",
			PortDescription: "Chooses the mipmap that most closely matches the size of the pixel being textured and uses the NearestFilter criterion (the texel nearest to the center of the pixel) to produce a texture value.",
		},

		"Linear Mipmap Nearest": nodes.ConstOutput[SamplerMinFilter]{
			Ref:             p,
			Val:             SamplerMinFilter_LINEAR_MIPMAP_NEAREST,
			PortName:        "Linear Mipmap Nearest",
			PortDescription: "Chooses the mipmap that most closely matches the size of the pixel being textured and uses the LinearFilter criterion (a weighted average of the four texels that are closest to the center of the pixel) to produce a texture value.",
		},

		"Nearest Mipmap Linear": nodes.ConstOutput[SamplerMinFilter]{
			Ref:             p,
			Val:             SamplerMinFilter_NEAREST_MIPMAP_LINEAR,
			PortName:        "Nearest Mipmap Linear",
			PortDescription: "Chooses the two mipmaps that most closely match the size of the pixel being textured and uses the NearestFilter criterion to produce a texture value from each mipmap. The final texture value is a weighted average of those two values.",
		},

		"Linear Mipmap Linear": nodes.ConstOutput[SamplerMinFilter]{
			Ref:             p,
			Val:             SamplerMinFilter_LINEAR_MIPMAP_LINEAR,
			PortName:        "Linear Mipmap Linear",
			PortDescription: "Chooses the two mipmaps that most closely match the size of the pixel being textured and uses the LinearFilter criterion to produce a texture value from each mipmap. The final texture value is a weighted average of those two values.",
		},
	}
}

// https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html#reference-material

type MaterialNode struct {
	Color                    nodes.Output[coloring.Color]  `description:"The factors for the base color of the material. This value defines linear multipliers for the sampled texels of the base color texture."`
	ColorTexture             nodes.Output[PolyformTexture] `description:"The base color texture. The first three components (RGB) MUST be encoded with the sRGB transfer function. They specify the base color of the material. If the fourth component (A) is present, it represents the linear alpha coverage of the material. Otherwise, the alpha coverage is equal to 1.0. The material.alphaMode property specifies how alpha is interpreted. The stored texels MUST NOT be premultiplied. When undefined, the texture MUST be sampled as having 1.0 in all components."`
	MetallicFactor           nodes.Output[float64]         `description:"The factor for the metalness of the material. This value defines a linear multiplier for the sampled metalness values of the metallic-roughness texture."`
	RoughnessFactor          nodes.Output[float64]         `description:"The factor for the roughness of the material. This value defines a linear multiplier for the sampled roughness values of the metallic-roughness texture."`
	MetallicRoughnessTexture nodes.Output[PolyformTexture] `description:"The metallic-roughness texture. The metalness values are sampled from the B channel. The roughness values are sampled from the G channel. These values MUST be encoded with a linear transfer function. If other channels are present (R or A), they MUST be ignored for metallic-roughness calculations. When undefined, the texture MUST be sampled as having 1.0 in G and B components."`
	EmissiveTexture          nodes.Output[PolyformTexture] `description:"The emissive texture. It controls the color and intensity of the light being emitted by the material. This texture contains RGB components encoded with the sRGB transfer function. If a fourth component (A) is present, it MUST be ignored. When undefined, the texture MUST be sampled as having 1.0 in RGB components."`
	EmissiveFactor           nodes.Output[coloring.Color]  `description:"The factors for the emissive color of the material. This value defines linear multipliers for the sampled texels of the emissive texture."`
	NormalTexture            nodes.Output[PolyformNormal]  `description:"The tangent space normal texture. The texture encodes RGB components with linear transfer function. Each texel represents the XYZ components of a normal vector in tangent space. The normal vectors use the convention +X is right and +Y is up. +Z points toward the viewer. If a fourth component (A) is present, it **MUST** be ignored. When undefined, the material does not have a tangent space normal texture."`

	// Extensions
	IndexOfRefraction nodes.Output[float64]
	Transmission      nodes.Output[PolyformTransmission]
	Volume            nodes.Output[PolyformVolume]
	Anisotropy        nodes.Output[PolyformAnisotropy]
	Clearcoat         nodes.Output[PolyformClearcoat]
	EmissiveStrength  nodes.Output[float64]
}

func (gmnd MaterialNode) Out(out *nodes.StructOutput[PolyformMaterial]) {
	var pbr *PolyformPbrMetallicRoughness

	if gmnd.Color != nil {
		pbr = &PolyformPbrMetallicRoughness{}
		pbr.BaseColorFactor = nodes.GetOutputValue(out, gmnd.Color)
	}

	if gmnd.ColorTexture != nil {
		tex := nodes.GetOutputReference(out, gmnd.ColorTexture)
		if tex.canAddToGLTF() {
			if pbr == nil {
				pbr = &PolyformPbrMetallicRoughness{}
			}
			pbr.BaseColorTexture = tex
		}
	}

	var emissiveTexture *PolyformTexture
	if gmnd.EmissiveTexture != nil {
		tex := nodes.GetOutputReference(out, gmnd.EmissiveTexture)
		if tex.canAddToGLTF() {
			emissiveTexture = tex
		}
	}

	if gmnd.MetallicRoughnessTexture != nil {
		tex := nodes.GetOutputReference(out, gmnd.MetallicRoughnessTexture)
		if tex.canAddToGLTF() {
			if pbr == nil {
				pbr = &PolyformPbrMetallicRoughness{}
			}
			pbr.MetallicRoughnessTexture = tex
		}
	}

	if gmnd.MetallicFactor != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		pbr.MetallicFactor = nodes.GetOutputReference(out, gmnd.MetallicFactor)
	}

	if gmnd.RoughnessFactor != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		pbr.RoughnessFactor = nodes.GetOutputReference(out, gmnd.RoughnessFactor)
	}

	var emissiveFactor color.Color
	if gmnd.EmissiveFactor != nil {
		emissiveFactor = nodes.GetOutputValue(out, gmnd.EmissiveFactor)
	}

	extensions := make([]MaterialExtension, 0)
	if gmnd.Transmission != nil {
		extensions = append(extensions, nodes.GetOutputValue(out, gmnd.Transmission))
	}

	if gmnd.Volume != nil {
		extensions = append(extensions, nodes.GetOutputValue(out, gmnd.Volume))
	}

	if gmnd.IndexOfRefraction != nil {
		extensions = append(extensions, PolyformIndexOfRefraction{
			IOR: nodes.GetOutputReference(out, gmnd.IndexOfRefraction),
		})
	}

	if gmnd.Anisotropy != nil {
		extensions = append(extensions, nodes.GetOutputValue(out, gmnd.Anisotropy))
	}

	if gmnd.Clearcoat != nil {
		extensions = append(extensions, nodes.GetOutputValue(out, gmnd.Clearcoat))
	}

	if gmnd.EmissiveStrength != nil {
		extensions = append(extensions, PolyformEmissiveStrength{
			EmissiveStrength: nodes.GetOutputReference(out, gmnd.EmissiveStrength),
		})
	}

	var normalTex *PolyformNormal
	if gmnd.NormalTexture != nil {
		tex := nodes.TryGetOutputReference(out, gmnd.NormalTexture, nil)
		if tex.canAddToGLTF() {
			normalTex = tex
		}
	}

	out.Set(PolyformMaterial{
		PbrMetallicRoughness: pbr,
		Extensions:           extensions,
		EmissiveFactor:       emissiveFactor,
		NormalTexture:        normalTex,
		EmissiveTexture:      emissiveTexture,
	})
}

func (gmnd MaterialNode) Description() string {
	return "The material appearance of a primitive"
}

type MaterialTransmissionExtensionNode struct {
	Factor  nodes.Output[float64]         `description:"The base percentage of light that is transmitted through the surface"`
	Texture nodes.Output[PolyformTexture] `description:"A texture that defines the transmission percentage of the surface, stored in the R channel. This will be multiplied by transmissionFactor."`
}

func (node MaterialTransmissionExtensionNode) Out(out *nodes.StructOutput[PolyformTransmission]) {
	out.Set(PolyformTransmission{
		Factor:  nodes.TryGetOutputValue(out, node.Factor, 0.),
		Texture: nodes.TryGetOutputReference(out, node.Texture, nil),
	})
}

func (node MaterialTransmissionExtensionNode) Description() string {
	return "The KHR_materials_transmission extension provides a way to define glTF 2.0 materials that are transparent to light in a physically plausible way. That is, it enables the creation of transparent materials that absorb, reflect and transmit light depending on the incident angle and the wavelength of light. Common uses cases for thin-surface transmissive materials include plastics and glass."
}

type MaterialVolumeExtensionNode struct {
	ThicknessFactor     nodes.Output[float64]        `description:"The thickness of the volume beneath the surface. The value is given in the coordinate space of the mesh. If the value is 0 the material is thin-walled. Otherwise the material is a volume boundary. The doubleSided property has no effect on volume boundaries. Range is [0, +inf)."`
	AttenuationDistance nodes.Output[float64]        `description:"Density of the medium given as the average distance that light travels in the medium before interacting with a particle. The value is given in world space. Range is (0, +inf)."`
	AttenuationColor    nodes.Output[coloring.Color] `description:"The color that white light turns into due to absorption when reaching the attenuation distance."`
}

func (node MaterialVolumeExtensionNode) Out(out *nodes.StructOutput[PolyformVolume]) {
	out.Set(PolyformVolume{
		ThicknessFactor:     nodes.TryGetOutputValue(out, node.ThicknessFactor, 0),
		AttenuationColor:    nodes.TryGetOutputValue(out, node.AttenuationColor, coloring.White()),
		AttenuationDistance: nodes.TryGetOutputReference(out, node.AttenuationDistance, nil),
	})
}

func (node MaterialVolumeExtensionNode) Description() string {
	return "By default, a glTF 2.0 material describes the scattering properties of a surface enclosing an infinitely thin volume. The surface defined by the mesh represents a thin wall. The volume extension makes it possible to turn the surface into an interface between volumes. The mesh to which the material is attached defines the boundaries of an homogeneous medium and therefore must be manifold. Volumes provide effects like refraction, absorption and scattering. Scattering is not subject of this extension."
}

type MaterialAnisotropyExtensionNode struct {
	AnisotropyStrength nodes.Output[float64] `description:"The anisotropy strength. When the anisotropy texture is present, this value is multiplied by the texture's blue channel."`
	AnisotropyRotation nodes.Output[float64] `description:"The rotation of the anisotropy in tangent, bitangent space, measured in radians counter-clockwise from the tangent. When the anisotropy texture is present, this value provides additional rotation to the vectors in the texture."`
}

func (node MaterialAnisotropyExtensionNode) Out(out *nodes.StructOutput[PolyformAnisotropy]) {
	out.Set(PolyformAnisotropy{
		AnisotropyStrength: nodes.TryGetOutputValue(out, node.AnisotropyStrength, 0),
		AnisotropyRotation: nodes.TryGetOutputValue(out, node.AnisotropyRotation, 0),
	})
}

func (node MaterialAnisotropyExtensionNode) Description() string {
	return "This extension defines the anisotropic property of a material as observable with brushed metals for example. An asymmetric specular lobe model is introduced to allow for such phenomena. The visually distinct feature of that lobe is the elongated appearance of the specular reflection."
}

type MaterialClearcoatExtensionNode struct {
	ClearcoatFactor          nodes.Output[float64]
	ClearcoatRoughnessFactor nodes.Output[float64]
}

func (node MaterialClearcoatExtensionNode) Out(out *nodes.StructOutput[PolyformClearcoat]) {
	out.Set(PolyformClearcoat{
		ClearcoatFactor:          nodes.TryGetOutputValue(out, node.ClearcoatFactor, 0),
		ClearcoatRoughnessFactor: nodes.TryGetOutputValue(out, node.ClearcoatRoughnessFactor, 0),
	})
}

func (node MaterialClearcoatExtensionNode) Description() string {
	return "A clear coat is a common technique used in Physically-Based Rendering to represent a protective layer applied to a base material."
}

type AnimationNode struct {
	Name     nodes.Output[string]
	Channels []nodes.Output[PolyformAnimationChannel]
}

func (node AnimationNode) Out(out *nodes.StructOutput[PolyformAnimation]) {
	out.Set(PolyformAnimation{
		Name:     nodes.TryGetOutputValue(out, node.Name, ""),
		Channels: nodes.GetOutputValues(out, node.Channels),
	})
}

func validateChannelNode[T any](
	out nodes.ExecutionRecorder,
	Target nodes.Output[*PolyformModel],
	Frames nodes.Output[[]animation.Frame[T]],
) (*PolyformModel, []animation.Frame[T], error) {
	target := nodes.TryGetOutputValue(out, Target, nil)
	if target == nil {
		return nil, nil, nodes.NilInputError{Input: Target}
	}

	frames := nodes.TryGetOutputValue(out, Frames, nil)
	if len(frames) == 0 {
		return nil, nil, nodes.InvalidInputError{
			Input:   Frames,
			Message: "Can't create an animation with 0 frames",
		}
	}

	return target, frames, nil
}

type TranslationAnimationChannelNode struct {
	Target        nodes.Output[*PolyformModel]
	Interpolation nodes.Output[AnimationSamplerInterpolation]
	Frames        nodes.Output[[]animation.Frame[vector3.Float64]]
}

func (node TranslationAnimationChannelNode) Out(out *nodes.StructOutput[PolyformAnimationChannel]) {
	target, frames, err := validateChannelNode(out, node.Target, node.Frames)
	if err != nil {
		out.CaptureError(err)
		return
	}

	times := make([]float64, len(frames))
	for i, v := range frames {
		times[i] = v.Time()
	}

	interpolation := nodes.TryGetOutputValue(out, node.Interpolation, AnimationSamplerInterpolation_LINEAR)

	var data []vector3.Float64
	switch interpolation {
	case AnimationSamplerInterpolation_LINEAR, AnimationSamplerInterpolation_STEP:
		data = make([]vector3.Float64, len(times))
		for i, v := range frames {
			data[i] = v.Val()
		}

	default:
		out.CaptureError(nodes.InvalidInputError{
			Input:   node.Interpolation,
			Message: "unimplemented interpolation",
		})
		return
	}

	out.Set(PolyformAnimationChannel{
		TargetPath: AnimationChannelTargetPath_TRANSLATION,
		Target:     target,
		Sampler: PolyformAnimationSampler{
			Interpolation: interpolation,
			Times:         times,
			Data:          Vector3AnimationSamplerData(data),
		},
	})
}

type RotationAnimationChannelNode struct {
	Target        nodes.Output[*PolyformModel]
	Interpolation nodes.Output[AnimationSamplerInterpolation]
	Frames        nodes.Output[[]animation.Frame[quaternion.Quaternion]]
}

func (node RotationAnimationChannelNode) Out(out *nodes.StructOutput[PolyformAnimationChannel]) {
	target, frames, err := validateChannelNode(out, node.Target, node.Frames)
	if err != nil {
		out.CaptureError(err)
		return
	}

	times := make([]float64, len(frames))
	for i, v := range frames {
		times[i] = v.Time()
	}

	interpolation := nodes.TryGetOutputValue(out, node.Interpolation, AnimationSamplerInterpolation_LINEAR)

	var data []quaternion.Quaternion
	switch interpolation {
	case AnimationSamplerInterpolation_LINEAR, AnimationSamplerInterpolation_STEP:
		data = make([]quaternion.Quaternion, len(times))
		for i, v := range frames {
			data[i] = v.Val()
		}

	default:
		out.CaptureError(nodes.InvalidInputError{
			Input:   node.Interpolation,
			Message: "unimplemented interpolation",
		})
		return
	}

	out.Set(PolyformAnimationChannel{
		TargetPath: AnimationChannelTargetPath_ROTATION,
		Target:     target,
		Sampler: PolyformAnimationSampler{
			Interpolation: interpolation,
			Times:         times,
			Data:          RotationAnimationSamplerData(data),
		},
	})
}

type ScaleAnimationChannelNode struct {
	Target        nodes.Output[*PolyformModel]
	Interpolation nodes.Output[AnimationSamplerInterpolation]
	Frames        nodes.Output[[]animation.Frame[vector3.Float64]]
}

func (node ScaleAnimationChannelNode) Out(out *nodes.StructOutput[PolyformAnimationChannel]) {
	target, frames, err := validateChannelNode(out, node.Target, node.Frames)
	if err != nil {
		out.CaptureError(err)
		return
	}

	times := make([]float64, len(frames))
	for i, v := range frames {
		times[i] = v.Time()
	}

	interpolation := nodes.TryGetOutputValue(out, node.Interpolation, AnimationSamplerInterpolation_LINEAR)

	var data []vector3.Float64
	switch interpolation {
	case AnimationSamplerInterpolation_LINEAR, AnimationSamplerInterpolation_STEP:
		data = make([]vector3.Float64, len(times))
		for i, v := range frames {
			data[i] = v.Val()
		}

	default:
		out.CaptureError(nodes.InvalidInputError{
			Input:   node.Interpolation,
			Message: "unimplemented interpolation",
		})
		return
	}

	out.Set(PolyformAnimationChannel{
		TargetPath: AnimationChannelTargetPath_SCALE,
		Target:     target,
		Sampler: PolyformAnimationSampler{
			Interpolation: interpolation,
			Times:         times,
			Data:          Vector3AnimationSamplerData(data),
		},
	})
}
