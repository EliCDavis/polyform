package gltf

import (
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

	refutil.RegisterType[ArtifactNode](factory)
	refutil.RegisterType[MaterialAnisotropyExtensionNode](factory)
	refutil.RegisterType[MaterialClearcoatExtensionNode](factory)
	refutil.RegisterType[MaterialNode](factory)
	refutil.RegisterType[MaterialTransmissionExtensionNode](factory)
	refutil.RegisterType[MaterialVolumeExtensionNode](factory)
	refutil.RegisterType[ModelNode](factory)
	refutil.RegisterType[TextureNode](factory)

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

type ArtifactNode = nodes.Struct[ArtifactNodeData]

type ArtifactNodeData struct {
	Models []nodes.Output[PolyformModel]
}

func (gad ArtifactNodeData) Out() nodes.StructOutput[manifest.Artifact] {
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

	return nodes.NewStructOutput[manifest.Artifact](&Artifact{
		Scene: PolyformScene{
			Models: models,
		},
	})
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
	model := PolyformModel{Name: "Mesh"}

	if gmnd.Material != nil {
		v := gmnd.Material.Value()
		model.Material = &v
	}

	if gmnd.Mesh != nil {
		mesh := gmnd.Mesh.Value()
		model.Mesh = &mesh
	}

	if gmnd.GpuInstances != nil {
		model.GpuInstances = gmnd.GpuInstances.Value()
	}

	if gmnd.Translation != nil {
		v := gmnd.Translation.Value()
		model.Translation = &v
	}

	if gmnd.Scale != nil {
		v := gmnd.Scale.Value()
		model.Scale = &v
	}

	if gmnd.Rotation != nil {
		v := gmnd.Rotation.Value()
		model.Rotation = &v
	}

	return nodes.NewStructOutput(model)
}

type TextureNode = nodes.Struct[TextureNodeData]

type TextureNodeData struct {
	URI nodes.Output[string]
}

func (tnd TextureNodeData) Out() nodes.StructOutput[PolyformTexture] {
	tex := PolyformTexture{}

	if tnd.URI != nil {
		tex.URI = tnd.URI.Value()
	}

	return nodes.NewStructOutput(tex)
}

func (gmnd TextureNodeData) Description() string {
	return "An object that combines an image and its sampler"
}

type MaterialNode = nodes.Struct[MaterialNodeData]

// https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html#reference-material

type MaterialNodeData struct {
	Color                    nodes.Output[coloring.WebColor] `description:"The factors for the base color of the material. This value defines linear multipliers for the sampled texels of the base color texture."`
	ColorTexture             nodes.Output[string]            `description:"The base color texture. The first three components (RGB) MUST be encoded with the sRGB transfer function. They specify the base color of the material. If the fourth component (A) is present, it represents the linear alpha coverage of the material. Otherwise, the alpha coverage is equal to 1.0. The material.alphaMode property specifies how alpha is interpreted. The stored texels MUST NOT be premultiplied. When undefined, the texture MUST be sampled as having 1.0 in all components."`
	MetallicFactor           nodes.Output[float64]           `description:"The factor for the metalness of the material. This value defines a linear multiplier for the sampled metalness values of the metallic-roughness texture."`
	RoughnessFactor          nodes.Output[float64]           `description:"The factor for the roughness of the material. This value defines a linear multiplier for the sampled roughness values of the metallic-roughness texture."`
	MetallicRoughnessTexture nodes.Output[string]            `description:"The metallic-roughness texture. The metalness values are sampled from the B channel. The roughness values are sampled from the G channel. These values MUST be encoded with a linear transfer function. If other channels are present (R or A), they MUST be ignored for metallic-roughness calculations. When undefined, the texture MUST be sampled as having 1.0 in G and B components."`
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
		pbr.BaseColorTexture = &PolyformTexture{
			URI: gmnd.ColorTexture.Value(),
		}
	}

	if gmnd.MetallicRoughnessTexture != nil {
		if pbr == nil {
			pbr = &PolyformPbrMetallicRoughness{}
		}
		pbr.MetallicRoughnessTexture = &PolyformTexture{
			URI: gmnd.MetallicRoughnessTexture.Value(),
		}
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
