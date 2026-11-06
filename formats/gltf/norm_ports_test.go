package gltf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/math/norm"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func occlusionStrength(t *testing.T, strength nodes.Output[norm.UFloat64]) float64 {
	t.Helper()
	node := &nodes.Struct[gltf.OcclusionTextureNode]{Data: gltf.OcclusionTextureNode{Strength: strength}}
	out := nodes.GetNodeOutputPort[gltf.PolyformOcclusion](node, "Out").Value()
	require.NotNil(t, out.Strength)
	return *out.Strength
}

func TestOcclusionStrengthUnassignedIsItsDefault(t *testing.T) {
	assert.Equal(t, 1., occlusionStrength(t, nil))
}

func TestOcclusionStrengthFedNothingReadsAsZero(t *testing.T) {
	assert.Equal(t, 0., occlusionStrength(t, nodes.ConstOutput[norm.UFloat64]{}))
}

func TestOcclusionStrengthFedAnEmptyParameterReadsAsZero(t *testing.T) {
	param := &parameter.UFloat64{}
	assert.Equal(t, 0., occlusionStrength(t, param.Outputs()["Value"].(nodes.Output[norm.UFloat64])))
}

func TestOcclusionStrengthOutOfRangeClamps(t *testing.T) {
	assert.Equal(t, 1., occlusionStrength(t, nodes.ConstOutput[norm.UFloat64]{Val: 5}))
}

func TestSheenRoughnessWithNoDefaultTagReadsAsZero(t *testing.T) {
	node := &nodes.Struct[gltf.MaterialSheenExtensionNode]{}
	out := nodes.GetNodeOutputPort[gltf.PolyformSheen](node, "Out").Value()
	assert.Equal(t, 0., out.SheenRoughnessFactor)
}
