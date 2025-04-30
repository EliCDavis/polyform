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
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ManifestNode](factory)
	refutil.RegisterType[MaterialAnisotropyExtensionNode](factory)
	refutil.RegisterType[MaterialClearcoatExtensionNode](factory)
	refutil.RegisterType[MaterialNode](factory)
	refutil.RegisterType[MaterialTransmissionExtensionNode](factory)
	refutil.RegisterType[MaterialVolumeExtensionNode](factory)
	refutil.RegisterType[ModelNode](factory)
	refutil.RegisterType[TextureReferenceNode](factory)
	refutil.RegisterType[TextureNode](factory)
	refutil.RegisterType[SamplerNode](factory)

	refutil.RegisterType[SamplerWrapNode](factory)
	refutil.RegisterType[SamplerMinFilterNode](factory)
	refutil.RegisterType[SamplerMagFilterNode](factory)

	generator.RegisterTypes(factory)
}

type Artifact struct {
	Scene PolyformScene
}

func (Artifact) Mime() string {
	return "model/gltf-binary"
}

func (ga Artifact) Write(w io.Writer) error {
	return WriteBinary(ga.Scene, w)
}

type ManifestNode = nodes.Struct[ManifestNodeData]

type ManifestNodeData struct {
	Models []nodes.Output[PolyformModel]
}

func (gad ManifestNodeData) Out() nodes.StructOutput[manifest.Manifest] {
	models := make([]PolyformModel, 0, len(gad.Models))

	for _, m := range gad.Models {
		if m == nil {
			continue
		}
		value := m.Value()

		// TechDebt: Skip nodes without meshes as at the moment it'll cause stuff
		// to error out
		if value.Mesh == nil {
			continue
		}

		models = append(models, value)
	}

	entry := manifest.Entry{
		Artifact: &Artifact{
			Scene: PolyformScene{
				Models:           models,
				UseGpuInstancing: true,
			},
		},
	}

	return nodes.NewStructOutput(manifest.SingleEntryManifest("model.glb", entry))
}

type ModelNode = nodes.Struct[ModelNodeData]

type ModelNodeData struct {
	Mesh     nodes.Output[modeling.Mesh]
	Material nodes.Output[PolyformMaterial]

	Translation nodes.Output[vector3.Float64]
	Rotation    nodes.Output[quaternion.Quaternion]
	Scale       nodes.Output[vector3.Float64]

	GpuInstances nodes.Output[[]trs.TRS]
}

func (gmnd ModelNodeData) Out() nodes.StructOutput[PolyformModel] {
	model := PolyformModel{
		Name:         "Mesh",
		GpuInstances: nodes.TryGetOutputValue(gmnd.GpuInstances, nil),
	}

	if gmnd.Material != nil {
		v := gmnd.Material.Value()
		model.Material = &v
	}

	if gmnd.Mesh != nil {
		mesh := gmnd.Mesh.Value()
		model.Mesh = &mesh
	}

	transform := trs.Identity()

	if gmnd.Translation != nil {
		v := gmnd.Translation.Value()
		transform = transform.Translate(v)
	}

	if gmnd.Scale != nil {
		v := gmnd.Scale.Value()
		transform = transform.SetScale(v)
	}

	if gmnd.Rotation != nil {
		v := gmnd.Rotation.Value()
		transform = transform.SetRotation(v)
	}

	model.TRS = &transform

	if gmnd.GpuInstances != nil {
		model.GpuInstances = gmnd.GpuInstances.Value()
	}

	return nodes.NewStructOutput(model)
}

type TextureReferenceNode = nodes.Struct[TextureReferenceNodeData]

type TextureReferenceNodeData struct {
	URI     nodes.Output[string]
	Sampler nodes.Output[Sampler]
}

func (tnd TextureReferenceNodeData) Out() nodes.StructOutput[PolyformTexture] {
	var sampler *Sampler = nil
	if tnd.Sampler != nil {
		v := tnd.Sampler.Value()
		sampler = &v
	}

	return nodes.NewStructOutput(PolyformTexture{
		Sampler: sampler,
		URI:     nodes.TryGetOutputValue(tnd.URI, ""),
	})
}

func (gmnd TextureReferenceNodeData) Description() string {
	return "An object that combines an image and its sampler"
}

type TextureNode = nodes.Struct[TextureNodeData]

type TextureNodeData struct {
	Image   nodes.Output[image.Image]
	Sampler nodes.Output[Sampler]
}

func (tnd TextureNodeData) Out() nodes.StructOutput[PolyformTexture] {
	var sampler *Sampler = nil
	if tnd.Sampler != nil {
		v := tnd.Sampler.Value()
		sampler = &v
	}

	return nodes.NewStructOutput(PolyformTexture{
		Sampler: sampler,
		Image:   nodes.TryGetOutputValue(tnd.Image, nil),
	})
}

func (gmnd TextureNodeData) Description() string {
	return "An object that combines an image and its sampler"
}

type SamplerNode = nodes.Struct[SamplerNodeData]

type SamplerNodeData struct {
	MagFilter nodes.Output[SamplerMagFilter] `description:"Magnification filter"`
	MinFilter nodes.Output[SamplerMinFilter] `description:"Minification filter"`
	WrapS     nodes.Output[SamplerWrap]      `description:"S (U) wrapping mode"`
	WrapT     nodes.Output[SamplerWrap]      `description:"T (V) wrapping mode"`
}

func (tnd SamplerNodeData) Out() nodes.StructOutput[Sampler] {
	return nodes.NewStructOutput(Sampler{
		MagFilter: nodes.TryGetOutputValue(tnd.MagFilter, SamplerMagFilter_NEAREST),
		MinFilter: nodes.TryGetOutputValue(tnd.MinFilter, SamplerMinFilter_NEAREST),
		WrapS:     nodes.TryGetOutputValue(tnd.WrapS, SamplerWrap_REPEAT),
		WrapT:     nodes.TryGetOutputValue(tnd.WrapT, SamplerWrap_REPEAT),
	})
}

func (SamplerNodeData) Description() string {
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

type MaterialNode = nodes.Struct[MaterialNodeData]

// https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html#reference-material

type MaterialNodeData struct {
	Color                    nodes.Output[coloring.WebColor] `description:"The factors for the base color of the material. This value defines linear multipliers for the sampled texels of the base color texture."`
	ColorTexture             nodes.Output[PolyformTexture]   `description:"The base color texture. The first three components (RGB) MUST be encoded with the sRGB transfer function. They specify the base color of the material. If the fourth component (A) is present, it represents the linear alpha coverage of the material. Otherwise, the alpha coverage is equal to 1.0. The material.alphaMode property specifies how alpha is interpreted. The stored texels MUST NOT be premultiplied. When undefined, the texture MUST be sampled as having 1.0 in all components."`
	MetallicFactor           nodes.Output[float64]           `description:"The factor for the metalness of the material. This value defines a linear multiplier for the sampled metalness values of the metallic-roughness texture."`
	RoughnessFactor          nodes.Output[float64]           `description:"The factor for the roughness of the material. This value defines a linear multiplier for the sampled roughness values of the metallic-roughness texture."`
	MetallicRoughnessTexture nodes.Output[PolyformTexture]   `description:"The metallic-roughness texture. The metalness values are sampled from the B channel. The roughness values are sampled from the G channel. These values MUST be encoded with a linear transfer function. If other channels are present (R or A), they MUST be ignored for metallic-roughness calculations. When undefined, the texture MUST be sampled as having 1.0 in G and B components."`
	EmissiveFactor           nodes.Output[coloring.WebColor] `description:"The factors for the emissive color of the material. This value defines linear multipliers for the sampled texels of the emissive texture."`

	// Extensions
	IndexOfRefraction nodes.Output[float64]
	Transmission      nodes.Output[PolyformTransmission]
	Volume            nodes.Output[PolyformVolume]
	Anisotropy        nodes.Output[PolyformAnisotropy]
	Clearcoat         nodes.Output[PolyformClearcoat]
	EmissiveStrength  nodes.Output[float64]
}

func (gmnd MaterialNodeData) Out() nodes.StructOutput[PolyformMaterial] {
	var pbr *PolyformPbrMetallicRoughness

	if gmnd.Color != nil {
		pbr = &PolyformPbrMetallicRoughness{}
		pbr.BaseColorFactor = gmnd.Color.Value()
	}

	if gmnd.ColorTexture != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		v := gmnd.ColorTexture.Value()
		pbr.BaseColorTexture = &v
	}

	if gmnd.MetallicRoughnessTexture != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		v := gmnd.MetallicRoughnessTexture.Value()
		pbr.MetallicRoughnessTexture = &v
	}

	if gmnd.MetallicFactor != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		v := gmnd.MetallicFactor.Value()
		pbr.MetallicFactor = &v
	}

	if gmnd.RoughnessFactor != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		v := gmnd.RoughnessFactor.Value()
		pbr.RoughnessFactor = &v
	}

	var emissiveFactor color.Color
	if gmnd.EmissiveFactor != nil {
		emissiveFactor = gmnd.EmissiveFactor.Value()
	}

	extensions := make([]MaterialExtension, 0)
	if gmnd.Transmission != nil {
		extensions = append(extensions, gmnd.Transmission.Value())
	}

	if gmnd.Volume != nil {
		extensions = append(extensions, gmnd.Volume.Value())
	}

	if gmnd.IndexOfRefraction != nil {
		v := gmnd.IndexOfRefraction.Value()
		extensions = append(extensions, PolyformIndexOfRefraction{
			IOR: &v,
		})
	}

	if gmnd.Anisotropy != nil {
		extensions = append(extensions, gmnd.Anisotropy.Value())
	}

	if gmnd.Clearcoat != nil {
		extensions = append(extensions, gmnd.Clearcoat.Value())
	}

	if gmnd.EmissiveStrength != nil {
		v := gmnd.EmissiveStrength.Value()
		extensions = append(extensions, PolyformEmissiveStrength{
			EmissiveStrength: &v,
		})
	}

	return nodes.NewStructOutput(PolyformMaterial{
		PbrMetallicRoughness: pbr,
		Extensions:           extensions,
		EmissiveFactor:       emissiveFactor,
	})
}

func (gmnd MaterialNodeData) Description() string {
	return "The material appearance of a primitive"
}

type MaterialTransmissionExtensionNode = nodes.Struct[MaterialTransmissionExtensionNodeData]

type MaterialTransmissionExtensionNodeData struct {
	Factor  nodes.Output[float64]         `description:"The base percentage of light that is transmitted through the surface"`
	Texture nodes.Output[PolyformTexture] `description:"A texture that defines the transmission percentage of the surface, stored in the R channel. This will be multiplied by transmissionFactor."`
}

func (node MaterialTransmissionExtensionNodeData) Out() nodes.StructOutput[PolyformTransmission] {
	transmission := PolyformTransmission{
		Factor: nodes.TryGetOutputValue(node.Factor, 0.),
	}

	if node.Texture != nil {
		v := node.Texture.Value()
		transmission.Texture = &v
	}

	return nodes.NewStructOutput(transmission)
}

func (node MaterialTransmissionExtensionNodeData) Description() string {
	return "The KHR_materials_transmission extension provides a way to define glTF 2.0 materials that are transparent to light in a physically plausible way. That is, it enables the creation of transparent materials that absorb, reflect and transmit light depending on the incident angle and the wavelength of light. Common uses cases for thin-surface transmissive materials include plastics and glass."
}

type MaterialVolumeExtensionNode = nodes.Struct[MaterialVolumeExtensionNodeData]

type MaterialVolumeExtensionNodeData struct {
	ThicknessFactor     nodes.Output[float64]           `description:"The thickness of the volume beneath the surface. The value is given in the coordinate space of the mesh. If the value is 0 the material is thin-walled. Otherwise the material is a volume boundary. The doubleSided property has no effect on volume boundaries. Range is [0, +inf)."`
	AttenuationDistance nodes.Output[float64]           `description:"Density of the medium given as the average distance that light travels in the medium before interacting with a particle. The value is given in world space. Range is (0, +inf)."`
	AttenuationColor    nodes.Output[coloring.WebColor] `description:"The color that white light turns into due to absorption when reaching the attenuation distance."`
}

func (node MaterialVolumeExtensionNodeData) Out() nodes.StructOutput[PolyformVolume] {
	var attenutationDistance *float64
	if node.AttenuationDistance != nil {
		v := node.AttenuationDistance.Value()
		attenutationDistance = &v
	}

	return nodes.NewStructOutput(PolyformVolume{
		ThicknessFactor:     nodes.TryGetOutputValue(node.ThicknessFactor, 0),
		AttenuationColor:    nodes.TryGetOutputValue(node.AttenuationColor, coloring.White()),
		AttenuationDistance: attenutationDistance,
	})
}

func (node MaterialVolumeExtensionNodeData) Description() string {
	return "By default, a glTF 2.0 material describes the scattering properties of a surface enclosing an infinitely thin volume. The surface defined by the mesh represents a thin wall. The volume extension makes it possible to turn the surface into an interface between volumes. The mesh to which the material is attached defines the boundaries of an homogeneous medium and therefore must be manifold. Volumes provide effects like refraction, absorption and scattering. Scattering is not subject of this extension."
}

type MaterialAnisotropyExtensionNode = nodes.Struct[MaterialAnisotropyExtensionNodeData]

type MaterialAnisotropyExtensionNodeData struct {
	AnisotropyStrength nodes.Output[float64] `description:"The anisotropy strength. When the anisotropy texture is present, this value is multiplied by the texture's blue channel."`
	AnisotropyRotation nodes.Output[float64] `description:"The rotation of the anisotropy in tangent, bitangent space, measured in radians counter-clockwise from the tangent. When the anisotropy texture is present, this value provides additional rotation to the vectors in the texture."`
}

func (node MaterialAnisotropyExtensionNodeData) Out() nodes.StructOutput[PolyformAnisotropy] {
	return nodes.NewStructOutput(PolyformAnisotropy{
		AnisotropyStrength: nodes.TryGetOutputValue(node.AnisotropyStrength, 0),
		AnisotropyRotation: nodes.TryGetOutputValue(node.AnisotropyRotation, 0),
	})
}

func (node MaterialAnisotropyExtensionNodeData) Description() string {
	return "This extension defines the anisotropic property of a material as observable with brushed metals for example. An asymmetric specular lobe model is introduced to allow for such phenomena. The visually distinct feature of that lobe is the elongated appearance of the specular reflection."
}

type MaterialClearcoatExtensionNode = nodes.Struct[MaterialClearcoatExtensionNodeData]

type MaterialClearcoatExtensionNodeData struct {
	ClearcoatFactor          nodes.Output[float64]
	ClearcoatRoughnessFactor nodes.Output[float64]
}

func (node MaterialClearcoatExtensionNodeData) Out() nodes.StructOutput[PolyformClearcoat] {
	return nodes.NewStructOutput(PolyformClearcoat{
		ClearcoatFactor:          nodes.TryGetOutputValue(node.ClearcoatFactor, 0),
		ClearcoatRoughnessFactor: nodes.TryGetOutputValue(node.ClearcoatRoughnessFactor, 0),
	})
}

func (node MaterialClearcoatExtensionNodeData) Description() string {
	return "A clear coat is a common technique used in Physically-Based Rendering to represent a protective layer applied to a base material."
}
