package gltf

import (
	"image/color"
	"io"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
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

type ArtifactNode = nodes.Struct[artifact.Artifact, ArtifactNodeData]

type ArtifactNodeData struct {
	Models []nodes.NodeOutput[PolyformModel]
}

func (gad ArtifactNodeData) Process() (artifact.Artifact, error) {
	models := make([]PolyformModel, 0, len(gad.Models))

	for _, m := range gad.Models {
		if m == nil {
			continue
		}
		models = append(models, m.Value())
	}

	return &Artifact{
		Scene: PolyformScene{
			Models: models,
		},
	}, nil
}

type ModelNode = nodes.Struct[PolyformModel, ModelNodeData]

type ModelNodeData struct {
	Mesh     nodes.NodeOutput[modeling.Mesh]
	Material nodes.NodeOutput[PolyformMaterial]

	Translation nodes.NodeOutput[vector3.Float64]
	Rotation    nodes.NodeOutput[quaternion.Quaternion]
	Scale       nodes.NodeOutput[vector3.Float64]

	GpuInstances nodes.NodeOutput[[]trs.TRS]
}

func (gmnd ModelNodeData) Process() (PolyformModel, error) {
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

	return model, nil
}

type TextureNode = nodes.Struct[PolyformTexture, TextureNodeData]

type TextureNodeData struct {
	URI nodes.NodeOutput[string]
}

func (tnd TextureNodeData) Process() (PolyformTexture, error) {
	tex := PolyformTexture{}

	if tnd.URI != nil {
		tex.URI = tnd.URI.Value()
	}

	return tex, nil
}

func (gmnd TextureNodeData) Description() string {
	return "An object that combines an image and its sampler"
}

type MaterialNode = nodes.Struct[PolyformMaterial, MaterialNodeData]

type MaterialNodeData struct {
	Color                    nodes.NodeOutput[coloring.WebColor]
	ColorTexture             nodes.NodeOutput[string]
	MetallicFactor           nodes.NodeOutput[float64]
	RoughnessFactor          nodes.NodeOutput[float64]
	MetallicRoughnessTexture nodes.NodeOutput[string]
	EmissiveFactor           nodes.NodeOutput[coloring.WebColor]

	// Extensions
	IndexOfRefraction nodes.NodeOutput[float64]
	Transmission      nodes.NodeOutput[PolyformTransmission]
	Volume            nodes.NodeOutput[PolyformVolume]
	Anisotropy        nodes.NodeOutput[PolyformAnisotropy]
	Clearcoat         nodes.NodeOutput[PolyformClearcoat]
	EmissiveStrength  nodes.NodeOutput[float64]
}

func (gmnd MaterialNodeData) Process() (PolyformMaterial, error) {
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

	return PolyformMaterial{
		PbrMetallicRoughness: pbr,
		Extensions:           extensions,
		EmissiveFactor:       emissiveFactor,
	}, nil
}

func (gmnd MaterialNodeData) Description() string {
	return "The material appearance of a primitive"
}

type MaterialTransmissionExtensionNode = nodes.Struct[PolyformTransmission, MaterialTransmissionExtensionNodeData]

type MaterialTransmissionExtensionNodeData struct {
	Factor  nodes.NodeOutput[float64]
	Texture nodes.NodeOutput[PolyformTexture]
}

func (gmvend MaterialTransmissionExtensionNodeData) Process() (PolyformTransmission, error) {
	transmission := PolyformTransmission{}

	if gmvend.Factor != nil {
		transmission.Factor = gmvend.Factor.Value()
	}

	if gmvend.Texture != nil {
		v := gmvend.Texture.Value()
		transmission.Texture = &v
	}

	return transmission, nil
}

func (gmvend MaterialTransmissionExtensionNodeData) Description() string {
	return "The KHR_materials_transmission extension provides a way to define glTF 2.0 materials that are transparent to light in a physically plausible way. That is, it enables the creation of transparent materials that absorb, reflect and transmit light depending on the incident angle and the wavelength of light. Common uses cases for thin-surface transmissive materials include plastics and glass."
}

type MaterialVolumeExtensionNode = nodes.Struct[PolyformVolume, MaterialVolumeExtensionNodeData]

type MaterialVolumeExtensionNodeData struct {
	ThicknessFactor     nodes.NodeOutput[float64]
	AttenuationDistance nodes.NodeOutput[float64]
	AttenuationColor    nodes.NodeOutput[coloring.WebColor]
}

func (gmvend MaterialVolumeExtensionNodeData) Process() (PolyformVolume, error) {
	var thickness float64
	var attenutationDistance *float64
	attenuationColor := coloring.White()

	if gmvend.ThicknessFactor != nil {
		thickness = gmvend.ThicknessFactor.Value()
	}

	if gmvend.AttenuationDistance != nil {
		v := gmvend.AttenuationDistance.Value()
		attenutationDistance = &v
	}

	if gmvend.AttenuationColor != nil {
		attenuationColor = gmvend.AttenuationColor.Value()
	}

	return PolyformVolume{
		ThicknessFactor:     thickness,
		AttenuationColor:    attenuationColor,
		AttenuationDistance: attenutationDistance,
	}, nil
}

func (gmvend MaterialVolumeExtensionNodeData) Description() string {
	return "By default, a glTF 2.0 material describes the scattering properties of a surface enclosing an infinitely thin volume. The surface defined by the mesh represents a thin wall. The volume extension makes it possible to turn the surface into an interface between volumes. The mesh to which the material is attached defines the boundaries of an homogeneous medium and therefore must be manifold. Volumes provide effects like refraction, absorption and scattering. Scattering is not subject of this extension."
}

type MaterialAnisotropyExtensionNode = nodes.Struct[PolyformAnisotropy, MaterialAnisotropyExtensionNodeData]

type MaterialAnisotropyExtensionNodeData struct {
	AnisotropyStrength nodes.NodeOutput[float64]
	AnisotropyRotation nodes.NodeOutput[float64]
}

func (gmvend MaterialAnisotropyExtensionNodeData) Process() (PolyformAnisotropy, error) {
	var strength float64
	var rotation float64

	if gmvend.AnisotropyStrength != nil {
		strength = gmvend.AnisotropyStrength.Value()
	}

	if gmvend.AnisotropyRotation != nil {
		rotation = gmvend.AnisotropyRotation.Value()
	}

	return PolyformAnisotropy{
		AnisotropyStrength: strength,
		AnisotropyRotation: rotation,
	}, nil
}

func (gmvend MaterialAnisotropyExtensionNodeData) Description() string {
	return "This extension defines the anisotropic property of a material as observable with brushed metals for example. An asymmetric specular lobe model is introduced to allow for such phenomena. The visually distinct feature of that lobe is the elongated appearance of the specular reflection."
}

type MaterialClearcoatExtensionNode = nodes.Struct[PolyformClearcoat, MaterialClearcoatExtensionNodeData]

type MaterialClearcoatExtensionNodeData struct {
	ClearcoatFactor          nodes.NodeOutput[float64]
	ClearcoatRoughnessFactor nodes.NodeOutput[float64]
}

func (gmcend MaterialClearcoatExtensionNodeData) Process() (PolyformClearcoat, error) {
	var strength float64
	var rotation float64

	if gmcend.ClearcoatFactor != nil {
		strength = gmcend.ClearcoatFactor.Value()
	}

	if gmcend.ClearcoatRoughnessFactor != nil {
		rotation = gmcend.ClearcoatRoughnessFactor.Value()
	}

	return PolyformClearcoat{
		ClearcoatFactor:          strength,
		ClearcoatRoughnessFactor: rotation,
	}, nil
}

func (gmcend MaterialClearcoatExtensionNodeData) Description() string {
	return "A clear coat is a common technique used in Physically-Based Rendering to represent a protective layer applied to a base material."
}
