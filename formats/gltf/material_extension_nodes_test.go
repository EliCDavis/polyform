package gltf_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/nodes"
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
}
