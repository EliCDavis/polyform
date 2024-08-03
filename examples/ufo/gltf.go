package main

import (
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
)

type GltfArtifact = nodes.StructNode[generator.Artifact, GltfArtifactData]

type GltfArtifactData struct {
	Models []nodes.NodeOutput[gltf.PolyformModel]
}

func (gad GltfArtifactData) Process() (generator.Artifact, error) {
	models := make([]gltf.PolyformModel, len(gad.Models))

	for i, m := range gad.Models {
		models[i] = m.Value()
	}

	return &artifact.Gltf{
		Scene: gltf.PolyformScene{
			Models: models,
		},
	}, nil
}

type GltfModel = nodes.StructNode[gltf.PolyformModel, GltfModelData]

type GltfModelData struct {
	Mesh     nodes.NodeOutput[modeling.Mesh]
	Material nodes.NodeOutput[gltf.PolyformMaterial]
}

func (gad GltfModelData) Process() (gltf.PolyformModel, error) {
	var mat *gltf.PolyformMaterial
	if gad.Material != nil {
		v := gad.Material.Value()
		mat = &v
	}

	return gltf.PolyformModel{
		Name:     "Mesh",
		Mesh:     gad.Mesh.Value(),
		Material: mat,
	}, nil
}

type GltfMaterialNode = nodes.StructNode[gltf.PolyformMaterial, GltfMaterialNodeData]

type GltfMaterialNodeData struct {
	Color           nodes.NodeOutput[coloring.WebColor]
	MetallicFactor  nodes.NodeOutput[float64]
	RoughnessFactor nodes.NodeOutput[float64]
	EmissiveFactor  nodes.NodeOutput[coloring.WebColor]

	// Extensions
	IndexOfRefraction nodes.NodeOutput[float64]
	Transmission      nodes.NodeOutput[gltf.PolyformTransmission]
	Volume            nodes.NodeOutput[gltf.PolyformVolume]
	Anisotropy        nodes.NodeOutput[gltf.PolyformAnisotropy]
	Clearcoat         nodes.NodeOutput[gltf.PolyformClearcoat]
	EmissiveStrength  nodes.NodeOutput[float64]
}

func (gmnd GltfMaterialNodeData) Process() (gltf.PolyformMaterial, error) {
	var pbr *gltf.PolyformPbrMetallicRoughness

	if gmnd.Color != nil {
		pbr = &gltf.PolyformPbrMetallicRoughness{}
		pbr.BaseColorFactor = gmnd.Color.Value()
	}

	if gmnd.MetallicFactor != nil {
		if pbr == nil {
			pbr = &gltf.PolyformPbrMetallicRoughness{}
		}
		v := gmnd.MetallicFactor.Value()
		pbr.MetallicFactor = &v
	}

	if gmnd.RoughnessFactor != nil {
		if pbr == nil {
			pbr = &gltf.PolyformPbrMetallicRoughness{}
		}
		v := gmnd.RoughnessFactor.Value()
		pbr.RoughnessFactor = &v
	}

	var emissiveFactor color.Color
	if gmnd.EmissiveFactor != nil {
		emissiveFactor = gmnd.EmissiveFactor.Value()
	}

	extensions := make([]gltf.MaterialExtension, 0)
	if gmnd.Transmission != nil {
		extensions = append(extensions, gmnd.Transmission.Value())
	}

	if gmnd.Volume != nil {
		extensions = append(extensions, gmnd.Volume.Value())
	}

	if gmnd.IndexOfRefraction != nil {
		v := gmnd.IndexOfRefraction.Value()
		extensions = append(extensions, gltf.PolyformIndexOfRefraction{
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
		extensions = append(extensions, gltf.PolyformEmissiveStrength{
			EmissiveStrength: &v,
		})
	}

	return gltf.PolyformMaterial{
		PbrMetallicRoughness: pbr,
		Extensions:           extensions,
		EmissiveFactor:       emissiveFactor,
	}, nil
}

type GltfMaterialTransmissionExtensionNode = nodes.StructNode[gltf.PolyformTransmission, GltfMaterialTransmissionExtensionNodeData]

type GltfMaterialTransmissionExtensionNodeData struct {
	TransmissionFactor nodes.NodeOutput[float64]
}

func (gmvend GltfMaterialTransmissionExtensionNodeData) Process() (gltf.PolyformTransmission, error) {
	var transmissionFactor float64

	if gmvend.TransmissionFactor != nil {
		transmissionFactor = gmvend.TransmissionFactor.Value()
	}

	return gltf.PolyformTransmission{
		Factor: transmissionFactor,
	}, nil
}

type GltfMaterialVolumeExtensionNode = nodes.StructNode[gltf.PolyformVolume, GltfMaterialVolumeExtensionNodeData]

type GltfMaterialVolumeExtensionNodeData struct {
	ThicknessFactor     nodes.NodeOutput[float64]
	AttenuationDistance nodes.NodeOutput[float64]
	AttenuationColor    nodes.NodeOutput[coloring.WebColor]
}

func (gmvend GltfMaterialVolumeExtensionNodeData) Process() (gltf.PolyformVolume, error) {
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

	return gltf.PolyformVolume{
		ThicknessFactor:     thickness,
		AttenuationColor:    attenuationColor,
		AttenuationDistance: attenutationDistance,
	}, nil
}

type GltfMaterialAnisotropyExtensionNode = nodes.StructNode[gltf.PolyformAnisotropy, GltfMaterialAnisotropyExtensionNodeData]

type GltfMaterialAnisotropyExtensionNodeData struct {
	AnisotropyStrength nodes.NodeOutput[float64]
	AnisotropyRotation nodes.NodeOutput[float64]
}

func (gmvend GltfMaterialAnisotropyExtensionNodeData) Process() (gltf.PolyformAnisotropy, error) {
	var strength float64
	var rotation float64

	if gmvend.AnisotropyStrength != nil {
		strength = gmvend.AnisotropyStrength.Value()
	}

	if gmvend.AnisotropyRotation != nil {
		rotation = gmvend.AnisotropyRotation.Value()
	}

	return gltf.PolyformAnisotropy{
		AnisotropyStrength: strength,
		AnisotropyRotation: rotation,
	}, nil
}

type GltfMaterialClearcoatExtensionNode = nodes.StructNode[gltf.PolyformClearcoat, GltfMaterialClearcoatExtensionNodeData]

type GltfMaterialClearcoatExtensionNodeData struct {
	ClearcoatFactor          nodes.NodeOutput[float64]
	ClearcoatRoughnessFactor nodes.NodeOutput[float64]
}

func (gmcend GltfMaterialClearcoatExtensionNodeData) Process() (gltf.PolyformClearcoat, error) {
	var strength float64
	var rotation float64

	if gmcend.ClearcoatFactor != nil {
		strength = gmcend.ClearcoatFactor.Value()
	}

	if gmcend.ClearcoatRoughnessFactor != nil {
		rotation = gmcend.ClearcoatRoughnessFactor.Value()
	}

	return gltf.PolyformClearcoat{
		ClearcoatFactor:          strength,
		ClearcoatRoughnessFactor: rotation,
	}, nil
}
