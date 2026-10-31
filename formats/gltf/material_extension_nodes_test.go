package gltf_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func materialWith(t *testing.T, set func(m *gltf.MaterialNode)) gltf.PolyformMaterial {
	t.Helper()
	data := gltf.MaterialNode{Color: nodes.ConstOutput[coloring.Color]{Val: coloring.White()}}
	set(&data)
	node := &nodes.Struct[gltf.MaterialNode]{Data: data}
	return nodes.GetNodeOutputPort[gltf.PolyformMaterial](node, "Out").Value()
}

func extensionIDs(m gltf.PolyformMaterial) []string {
	var ids []string
	for _, e := range m.Extensions {
		ids = append(ids, e.ExtensionID())
	}
	return ids
}

func glbJSON(t *testing.T, material gltf.PolyformMaterial) map[string]any {
	t.Helper()
	mesh := cube()
	buf := &bytes.Buffer{}
	require.NoError(t, gltf.WriteBinary(gltf.PolyformScene{Models: []*gltf.PolyformModel{{Name: "m", Mesh: &mesh, Material: &material}}}, buf, nil))
	raw := buf.Bytes()
	jsonLen := binary.LittleEndian.Uint32(raw[12:16])
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw[20:20+jsonLen], &doc))
	return doc
}

func TestMaterialNodeCarriesSheenSpecularIridescenceAndDispersion(t *testing.T) {
	sheen := nodes.GetNodeOutputPort[gltf.PolyformSheen](&nodes.Struct[gltf.MaterialSheenExtensionNode]{Data: gltf.MaterialSheenExtensionNode{
		ColorFactor:     nodes.ConstOutput[coloring.Color]{Val: coloring.Color{R: 0.8, G: 0.1, B: 0.1, A: 1}},
		RoughnessFactor: nodes.ConstOutput[float64]{Val: 0.6},
	}}, "Out").Value()
	specular := nodes.GetNodeOutputPort[gltf.PolyformSpecular](&nodes.Struct[gltf.MaterialSpecularExtensionNode]{Data: gltf.MaterialSpecularExtensionNode{
		Factor:      nodes.ConstOutput[float64]{Val: 0.5},
		ColorFactor: nodes.ConstOutput[coloring.Color]{Val: coloring.Color{R: 1, G: 0.9, B: 0.5, A: 1}},
	}}, "Out").Value()
	iridescence := nodes.GetNodeOutputPort[gltf.PolyformIridescence](&nodes.Struct[gltf.MaterialIridescenceExtensionNode]{Data: gltf.MaterialIridescenceExtensionNode{
		Factor:            nodes.ConstOutput[float64]{Val: 1},
		IndexOfRefraction: nodes.ConstOutput[float64]{Val: 1.5},
		ThicknessMinimum:  nodes.ConstOutput[float64]{Val: 200},
		ThicknessMaximum:  nodes.ConstOutput[float64]{Val: 600},
	}}, "Out").Value()
	dispersion := nodes.GetNodeOutputPort[gltf.PolyformDispersion](&nodes.Struct[gltf.MaterialDispersionExtensionNode]{Data: gltf.MaterialDispersionExtensionNode{
		Dispersion: nodes.ConstOutput[float64]{Val: 0.36},
	}}, "Out").Value()

	material := materialWith(t, func(m *gltf.MaterialNode) {
		m.Sheen = nodes.ConstOutput[gltf.PolyformSheen]{Val: sheen}
		m.Specular = nodes.ConstOutput[gltf.PolyformSpecular]{Val: specular}
		m.Iridescence = nodes.ConstOutput[gltf.PolyformIridescence]{Val: iridescence}
		m.Dispersion = nodes.ConstOutput[gltf.PolyformDispersion]{Val: dispersion}
	})
	assert.ElementsMatch(t, []string{"KHR_materials_sheen", "KHR_materials_specular", "KHR_materials_iridescence", "KHR_materials_dispersion"}, extensionIDs(material))

	doc := glbJSON(t, material)
	used, _ := doc["extensionsUsed"].([]any)
	assert.ElementsMatch(t, []any{"KHR_materials_sheen", "KHR_materials_specular", "KHR_materials_iridescence", "KHR_materials_dispersion"}, used)

	mat := doc["materials"].([]any)[0].(map[string]any)
	ext := mat["extensions"].(map[string]any)
	assert.InDelta(t, 0.6, ext["KHR_materials_sheen"].(map[string]any)["sheenRoughnessFactor"], 1e-9)
	assert.InDelta(t, 0.5, ext["KHR_materials_specular"].(map[string]any)["specularFactor"], 1e-9)
	assert.InDelta(t, 600, ext["KHR_materials_iridescence"].(map[string]any)["iridescenceThicknessMaximum"], 1e-9)
	assert.InDelta(t, 0.36, ext["KHR_materials_dispersion"].(map[string]any)["dispersion"], 1e-9)
}

func TestMaterialNodeWithoutTheNewPortsWritesNoneOfThem(t *testing.T) {
	material := materialWith(t, func(m *gltf.MaterialNode) {})
	assert.Empty(t, extensionIDs(material))
	assert.Nil(t, material.AlphaMode)
	assert.Nil(t, material.OcclusionTexture)
}

func texture(t *testing.T, data gltf.TextureNode) gltf.PolyformTexture {
	t.Helper()
	data.Image = nodes.ConstOutput[image.Image]{Val: image.NewRGBA(image.Rect(0, 0, 2, 2))}
	return nodes.GetNodeOutputPort[gltf.PolyformTexture](&nodes.Struct[gltf.TextureNode]{Data: data}, "Out").Value()
}

func textureOutput(t *testing.T) nodes.Output[gltf.PolyformTexture] {
	return nodes.ConstOutput[gltf.PolyformTexture]{Val: texture(t, gltf.TextureNode{})}
}

func TestMaterialNodeWritesAlphaModeOcclusionAndUnlit(t *testing.T) {
	occlusion := nodes.GetNodeOutputPort[gltf.PolyformOcclusion](&nodes.Struct[gltf.OcclusionTextureNode]{Data: gltf.OcclusionTextureNode{
		Texture:  textureOutput(t),
		Strength: nodes.ConstOutput[float64]{Val: 0.7},
	}}, "Out").Value()

	material := materialWith(t, func(m *gltf.MaterialNode) {
		m.AlphaMode = nodes.ConstOutput[gltf.AlphaMode]{Val: gltf.AlphaMask}
		m.AlphaCutoff = nodes.ConstOutput[float64]{Val: 0.3}
		m.Unlit = nodes.ConstOutput[bool]{Val: true}
		m.OcclusionTexture = nodes.ConstOutput[gltf.PolyformOcclusion]{Val: occlusion}
	})

	doc := glbJSON(t, material)
	mat := doc["materials"].([]any)[0].(map[string]any)
	assert.Equal(t, "MASK", mat["alphaMode"])
	assert.InDelta(t, 0.3, mat["alphaCutoff"], 1e-9)
	assert.InDelta(t, 0.7, mat["occlusionTexture"].(map[string]any)["strength"], 1e-9)
	assert.Contains(t, mat["extensions"], "KHR_materials_unlit")
	assert.Contains(t, doc["extensionsUsed"], "KHR_materials_unlit")
}

func TestMaterialNodeRefusesAStrayCutoff(t *testing.T) {
	for name, set := range map[string]func(m *gltf.MaterialNode){
		"cutoff with blend": func(m *gltf.MaterialNode) {
			m.AlphaMode = nodes.ConstOutput[gltf.AlphaMode]{Val: gltf.AlphaBlend}
			m.AlphaCutoff = nodes.ConstOutput[float64]{Val: 0.5}
		},
		"cutoff with the default opaque": func(m *gltf.MaterialNode) {
			m.AlphaCutoff = nodes.ConstOutput[float64]{Val: 0.5}
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := gltf.MaterialNode{}
			set(&data)
			node := &nodes.Struct[gltf.MaterialNode]{Data: data}
			material := nodes.GetNodeOutputPort[gltf.PolyformMaterial](node, "Out").Value()
			assert.Nil(t, material.AlphaCutoff)
			assert.NotEmpty(t, node.Outputs()["Out"].(nodes.ObservableExecution).ExecutionReport().Errors)
		})
	}
}

func TestExtensionNodesWriteTheirTextures(t *testing.T) {
	coatNormal := texture(t, gltf.TextureNode{})
	coatNormalScale := 0.4
	anisotropy := nodes.GetNodeOutputPort[gltf.PolyformAnisotropy](&nodes.Struct[gltf.MaterialAnisotropyExtensionNode]{Data: gltf.MaterialAnisotropyExtensionNode{
		AnisotropyStrength: nodes.ConstOutput[float64]{Val: 0.6},
		AnisotropyTexture:  textureOutput(t),
	}}, "Out").Value()
	volume := nodes.GetNodeOutputPort[gltf.PolyformVolume](&nodes.Struct[gltf.MaterialVolumeExtensionNode]{Data: gltf.MaterialVolumeExtensionNode{
		ThicknessFactor:  nodes.ConstOutput[float64]{Val: 0.1},
		ThicknessTexture: textureOutput(t),
	}}, "Out").Value()
	clearcoat := nodes.GetNodeOutputPort[gltf.PolyformClearcoat](&nodes.Struct[gltf.MaterialClearcoatExtensionNode]{Data: gltf.MaterialClearcoatExtensionNode{
		ClearcoatFactor:           nodes.ConstOutput[float64]{Val: 1},
		ClearcoatTexture:          textureOutput(t),
		ClearcoatRoughnessTexture: textureOutput(t),
		ClearcoatNormalTexture:    nodes.ConstOutput[gltf.PolyformNormal]{Val: gltf.PolyformNormal{PolyformTexture: &coatNormal, Scale: &coatNormalScale}},
	}}, "Out").Value()

	material := materialWith(t, func(m *gltf.MaterialNode) {
		m.Anisotropy = nodes.ConstOutput[gltf.PolyformAnisotropy]{Val: anisotropy}
		m.Volume = nodes.ConstOutput[gltf.PolyformVolume]{Val: volume}
		m.Clearcoat = nodes.ConstOutput[gltf.PolyformClearcoat]{Val: clearcoat}
	})

	ext := glbJSON(t, material)["materials"].([]any)[0].(map[string]any)["extensions"].(map[string]any)
	assert.Contains(t, ext["KHR_materials_anisotropy"], "anisotropyTexture")
	assert.Contains(t, ext["KHR_materials_volume"], "thicknessTexture")
	assert.Contains(t, ext["KHR_materials_clearcoat"], "clearcoatTexture")
	assert.Contains(t, ext["KHR_materials_clearcoat"], "clearcoatRoughnessTexture")
	coatNormalInfo := ext["KHR_materials_clearcoat"].(map[string]any)["clearcoatNormalTexture"].(map[string]any)
	assert.InDelta(t, 0.4, coatNormalInfo["scale"], 1e-9)
	assert.Contains(t, coatNormalInfo, "index")
}

func TestTextureNodeWritesATextureTransform(t *testing.T) {
	tex := texture(t, gltf.TextureNode{
		Scale:    nodes.ConstOutput[vector2.Float64]{Val: vector2.New(4., 2.)},
		Rotation: nodes.ConstOutput[float64]{Val: 0.5},
	})
	material := materialWith(t, func(m *gltf.MaterialNode) {
		m.ColorTexture = nodes.ConstOutput[gltf.PolyformTexture]{Val: tex}
	})

	doc := glbJSON(t, material)
	info := doc["materials"].([]any)[0].(map[string]any)["pbrMetallicRoughness"].(map[string]any)["baseColorTexture"].(map[string]any)
	transform := info["extensions"].(map[string]any)["KHR_texture_transform"].(map[string]any)
	assert.Equal(t, []any{4.0, 2.0}, transform["scale"])
	assert.InDelta(t, 0.5, transform["rotation"], 1e-9)
	assert.NotContains(t, transform, "offset")
	assert.Contains(t, doc["extensionsUsed"], "KHR_texture_transform")
}

func TestMaterialsDifferingOnlyInTheirImagesAreNotMerged(t *testing.T) {
	first := texture(t, gltf.TextureNode{})
	second := texture(t, gltf.TextureNode{})
	mesh := cube()
	scene := gltf.PolyformScene{Models: []*gltf.PolyformModel{
		{Name: "a", Mesh: &mesh, Material: &gltf.PolyformMaterial{Name: "Wood", PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{BaseColorTexture: &first}}},
		{Name: "b", Mesh: &mesh, Material: &gltf.PolyformMaterial{Name: "Wood", PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{BaseColorTexture: &second}}},
		{Name: "c", Mesh: &mesh, Material: &gltf.PolyformMaterial{Name: "Wood", NormalTexture: &gltf.PolyformNormal{PolyformTexture: &first}}},
	}}
	buf := &bytes.Buffer{}
	require.NoError(t, gltf.WriteBinary(scene, buf, nil))
	raw := buf.Bytes()
	jsonLen := binary.LittleEndian.Uint32(raw[12:16])
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw[20:20+jsonLen], &doc))
	assert.Len(t, doc["materials"], 3)
}
