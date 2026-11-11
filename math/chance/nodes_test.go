package chance_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/chance"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keys(n int) nodes.ConstOutput[[]int] {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return nodes.ConstOutput[[]int]{Val: out}
}

func floats(node nodes.Node, port string) []float64 {
	return nodes.GetNodeOutputPort[[]float64](node, port).Value()
}

func errorsOf(node nodes.Node, port string) []string {
	out := node.Outputs()[port]
	nodes.DynamicAnyValue(nil, out)
	return out.(nodes.ObservableExecution).ExecutionReport().Errors
}

func TestRandomNodeRegistersOncePerValueType(t *testing.T) {
	for _, name := range []string{
		"github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/chance.RandomNode[float64]]",
		"github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/chance.RandomNode[github.com/EliCDavis/vector/vector3.Vector[float64]]]",
		"github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/chance.RandomNode[github.com/EliCDavis/polyform/math/quaternion.Quaternion]]",
	} {
		assert.Contains(t, registry.Types().Types(), name)
	}
}

func TestRandomNodeLiftsOverKeys(t *testing.T) {
	node := &nodes.Struct[chance.RandomNode[float64]]{Data: chance.RandomNode[float64]{
		Key:          keys(4),
		Seed:         nodes.ConstOutput[int]{Val: 2},
		Distribution: nodes.ConstOutput[chance.Distribution[float64]]{Val: chance.Uniform(10, 20)},
	}}
	values := floats(node, "Out")
	require.Len(t, values, 4)
	for i, v := range values {
		assert.InDelta(t, 10+10*chance.Unit(2, i), v, 1e-9)
	}
}

func TestRandomNodeWithOneKeyGivesOneValue(t *testing.T) {
	node := &nodes.Struct[chance.RandomNode[float64]]{Data: chance.RandomNode[float64]{
		Key:          nodes.ConstOutput[int]{Val: 9},
		Distribution: nodes.ConstOutput[chance.Distribution[float64]]{Val: chance.Uniform(0, 1)},
	}}
	out := node.Outputs()["Out"]
	assert.Equal(t, "float64", out.(nodes.Typed).Type())
	assert.Equal(t, chance.Unit(0, 9), out.(nodes.Output[float64]).Value())
}

func TestRandomNodeNeedsADistribution(t *testing.T) {
	node := &nodes.Struct[chance.RandomNode[float64]]{Data: chance.RandomNode[float64]{Key: keys(3)}}
	assert.NotEmpty(t, errorsOf(node, "Out"))
}

func TestRandomNodeDrawsVectors(t *testing.T) {
	node := &nodes.Struct[chance.RandomNode[vector3.Float64]]{Data: chance.RandomNode[vector3.Float64]{
		Key:          keys(3),
		Distribution: nodes.ConstOutput[chance.Distribution[vector3.Float64]]{Val: chance.InsideSphere(1)},
	}}
	values := nodes.GetNodeOutputPort[[]vector3.Float64](node, "Out").Value()
	require.Len(t, values, 3)
	for _, v := range values {
		assert.LessOrEqual(t, v.Length(), 1.0+1e-9)
	}
}

func TestRandomNodesSharingSeedAndKeyDrawTogether(t *testing.T) {
	height := &nodes.Struct[chance.RandomNode[float64]]{Data: chance.RandomNode[float64]{
		Key: keys(50), Seed: nodes.ConstOutput[int]{Val: 1},
		Distribution: nodes.ConstOutput[chance.Distribution[float64]]{Val: chance.Uniform(0, 1)},
	}}
	width := &nodes.Struct[chance.RandomNode[float64]]{Data: chance.RandomNode[float64]{
		Key: keys(50), Seed: nodes.ConstOutput[int]{Val: 1},
		Distribution: nodes.ConstOutput[chance.Distribution[float64]]{Val: chance.Uniform(0, 100)},
	}}
	h, w := floats(height, "Out"), floats(width, "Out")
	for i := range h {
		assert.InDelta(t, h[i]*100, w[i], 1e-9)
	}
}

func TestRandomBoolFollowsProbability(t *testing.T) {
	node := &nodes.Struct[chance.RandomBoolNode]{Data: chance.RandomBoolNode{
		Key:         keys(10000),
		Probability: nodes.ConstOutput[float64]{Val: 0.2},
	}}
	trues := 0
	for _, b := range nodes.GetNodeOutputPort[[]bool](node, "Out").Value() {
		if b {
			trues++
		}
	}
	assert.InDelta(t, 2000, trues, 150)
}

func TestRandomIndexFollowsTheWeights(t *testing.T) {
	node := &nodes.Struct[chance.RandomIndexNode]{Data: chance.RandomIndexNode{
		Key:     keys(8000),
		Weights: nodes.ConstOutput[[]float64]{Val: []float64{3, 0, 1}},
	}}
	counts := make([]int, 3)
	for _, i := range nodes.GetNodeOutputPort[[]int](node, "Out").Value() {
		counts[i]++
	}
	assert.InDelta(t, 6000, counts[0], 250)
	assert.Zero(t, counts[1])
	assert.InDelta(t, 2000, counts[2], 250)
}

func TestRandomIndexRefusesUnusableWeights(t *testing.T) {
	for name, weights := range map[string][]float64{"none": nil, "all zero": {0, 0}, "negative": {1, -1}} {
		t.Run(name, func(t *testing.T) {
			node := &nodes.Struct[chance.RandomIndexNode]{Data: chance.RandomIndexNode{
				Key:     keys(3),
				Weights: nodes.ConstOutput[[]float64]{Val: weights},
			}}
			assert.NotEmpty(t, errorsOf(node, "Out"))
		})
	}
}

func TestSampleAppliesTheDistributionToGivenUnits(t *testing.T) {
	node := &nodes.Struct[chance.SampleNode]{Data: chance.SampleNode{
		Distribution: nodes.ConstOutput[chance.Distribution[float64]]{Val: chance.Uniform(0, 100)},
		Fraction:     nodes.ConstOutput[[]float64]{Val: []float64{0, 0.25, 1}},
	}}
	assert.Equal(t, []float64{0, 25, 100}, floats(node, "Out"))
}

func TestSeedNodeDerivesPerLabel(t *testing.T) {
	node := &nodes.Struct[chance.SeedNode]{Data: chance.SeedNode{
		Seed:  nodes.ConstOutput[int]{Val: 4},
		Label: nodes.ConstOutput[[]string]{Val: []string{"stools", "bottles"}},
	}}
	seeds := nodes.GetNodeOutputPort[[]int](node, "Out").Value()
	require.Len(t, seeds, 2)
	assert.Equal(t, chance.Derive(4, "stools"), seeds[0])
	assert.NotEqual(t, seeds[0], seeds[1])
}

func TestDistributionNodesBuildTheirDistribution(t *testing.T) {
	uniform := &nodes.Struct[chance.UniformNode]{Data: chance.UniformNode{
		Min: nodes.ConstOutput[float64]{Val: 1},
		Max: nodes.ConstOutput[float64]{Val: 3},
	}}
	d := nodes.GetNodeOutputPort[chance.Distribution[float64]](uniform, "Out").Value()
	assert.Equal(t, 2.0, at(d, 0.5))

	discrete := &nodes.Struct[chance.DiscreteNode]{Data: chance.DiscreteNode{
		Values:  nodes.ConstOutput[[]float64]{Val: []float64{1}},
		Weights: nodes.ConstOutput[[]float64]{Val: []float64{1, 2}},
	}}
	assert.NotEmpty(t, errorsOf(discrete, "Out"))

	compose := &nodes.Struct[chance.Compose3DNode]{Data: chance.Compose3DNode{
		X: nodes.ConstOutput[chance.Distribution[float64]]{Val: chance.Constant(1)},
	}}
	v := nodes.GetNodeOutputPort[chance.Distribution[vector3.Float64]](compose, "Out").Value()(chance.Fixed(0.3))
	assert.Equal(t, vector3.New(1.0, 0.3, 0.3), v, "an unwired axis stays unit uniform")
}
