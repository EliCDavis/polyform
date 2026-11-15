package gltf_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sceneJSON(t *testing.T, scene gltf.PolyformScene) map[string]any {
	t.Helper()
	buf := &bytes.Buffer{}
	require.NoError(t, gltf.WriteBinary(scene, buf, nil))
	raw := buf.Bytes()
	jsonLen := binary.LittleEndian.Uint32(raw[12:16])
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw[20:20+jsonLen], &doc))
	return doc
}

func lightOf(t *testing.T, set func(n *gltf.LightNode)) (gltf.KHR_LightsPunctual, []string) {
	t.Helper()
	data := gltf.LightNode{}
	set(&data)
	node := &nodes.Struct[gltf.LightNode]{Data: data}
	light := nodes.GetNodeOutputPort[gltf.KHR_LightsPunctual](node, "Out").Value()
	return light, node.Outputs()["Out"].(nodes.ObservableExecution).ExecutionReport().Errors
}

func TestModelLightIsWrittenOnTheModelsOwnNode(t *testing.T) {
	intensity := 3.0
	position := trs.Position(vector3.New(1., 2., 3.))
	doc := sceneJSON(t, gltf.PolyformScene{Models: []*gltf.PolyformModel{{
		Name: "lamp",
		TRS:  &position,
		Light: &gltf.KHR_LightsPunctual{
			Type:      gltf.KHR_LightsPunctualType_Point,
			Intensity: &intensity,
		},
	}}})

	node := doc["nodes"].([]any)[0].(map[string]any)
	assert.Equal(t, "lamp", node["name"])
	assert.Equal(t, []any{1., 2., 3.}, node["translation"])
	assert.Equal(t, map[string]any{"light": 0.}, node["extensions"].(map[string]any)["KHR_lights_punctual"])

	lights := doc["extensions"].(map[string]any)["KHR_lights_punctual"].(map[string]any)["lights"].([]any)
	require.Len(t, lights, 1)
	assert.Equal(t, "point", lights[0].(map[string]any)["type"])
	assert.Equal(t, 3., lights[0].(map[string]any)["intensity"])
	assert.Contains(t, doc["extensionsUsed"], "KHR_lights_punctual")
}

func TestChildModelLightsRepeatWithEachParent(t *testing.T) {
	bulb := &gltf.PolyformModel{Name: "bulb", Light: &gltf.KHR_LightsPunctual{Type: gltf.KHR_LightsPunctualType_Point}}
	a := trs.Position(vector3.New(-1., 0., 0.))
	b := trs.Position(vector3.New(1., 0., 0.))
	doc := sceneJSON(t, gltf.PolyformScene{Models: []*gltf.PolyformModel{
		{Name: "a", TRS: &a, Children: []*gltf.PolyformModel{bulb}},
		{Name: "b", TRS: &b, Children: []*gltf.PolyformModel{bulb}},
	}})

	lit := 0
	for _, n := range doc["nodes"].([]any) {
		if ext, ok := n.(map[string]any)["extensions"].(map[string]any); ok && ext["KHR_lights_punctual"] != nil {
			lit++
		}
	}
	assert.Equal(t, 2, lit)
}

func TestSpotLightWritesItsCone(t *testing.T) {
	inner, outer := 0.2, 0.6
	doc := sceneJSON(t, gltf.PolyformScene{Models: []*gltf.PolyformModel{{
		Name:  "spot",
		Light: &gltf.KHR_LightsPunctual{Type: gltf.KHR_LightsPunctualType_Spot, InnerConeAngle: &inner, OuterConeAngle: &outer},
	}}})

	light := doc["extensions"].(map[string]any)["KHR_lights_punctual"].(map[string]any)["lights"].([]any)[0].(map[string]any)
	assert.Equal(t, "spot", light["type"])
	assert.Equal(t, map[string]any{"innerConeAngle": 0.2, "outerConeAngle": 0.6}, light["spot"])
}

func TestLightNodeWritesPickedColorAsLinear(t *testing.T) {
	light, err := lightOf(t, func(n *gltf.LightNode) {
		n.Color = nodes.ConstOutput[coloring.Color]{Val: coloring.Color{R: 0.5, G: 0.5, B: 0.5, A: 1}}
		n.Intensity = nodes.ConstOutput[float64]{Val: 4}
		n.Range = nodes.ConstOutput[float64]{Val: 6}
	})
	require.Empty(t, err)

	r, _, _, _ := light.Color.RGBA()
	assert.InDelta(t, coloring.SRGBToLinear(0.5), float64(r)/math.MaxUint16, 1e-3)
	assert.Equal(t, 4., *light.Intensity)
	assert.Equal(t, 6., *light.Range)
	assert.Equal(t, gltf.KHR_LightsPunctualType_Point, light.Type)
}

func TestLightNodeLeavesRangeInfiniteAtZero(t *testing.T) {
	light, err := lightOf(t, func(n *gltf.LightNode) {})
	require.Empty(t, err)
	assert.Nil(t, light.Range)
}

func TestLightNodeRejectsAnInnerConeWiderThanTheOuter(t *testing.T) {
	_, err := lightOf(t, func(n *gltf.LightNode) {
		n.Type = nodes.ConstOutput[gltf.LightType]{Val: gltf.LightTypeSpot}
		n.InnerConeAngle = nodes.ConstOutput[float64]{Val: 0.8}
		n.OuterConeAngle = nodes.ConstOutput[float64]{Val: 0.4}
	})
	assert.NotEmpty(t, err)
}

func TestModelNodeCarriesItsLight(t *testing.T) {
	light := gltf.KHR_LightsPunctual{Type: gltf.KHR_LightsPunctualType_Point}
	node := &nodes.Struct[gltf.ModelNode]{Data: gltf.ModelNode{
		Light: nodes.ConstOutput[gltf.KHR_LightsPunctual]{Val: light},
	}}
	model := nodes.GetNodeOutputPort[*gltf.PolyformModel](node, "Out").Value()
	require.NotNil(t, model.Light)
	assert.Equal(t, gltf.KHR_LightsPunctualType_Point, model.Light.Type)
}
