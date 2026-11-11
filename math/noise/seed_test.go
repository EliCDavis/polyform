package noise_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func perlin3(seed int) []float64 {
	points := make([]vector3.Float64, 50)
	for i := range points {
		points[i] = vector3.New(float64(i)*0.37, 0.5, 1.1)
	}
	node := nodes.Struct[noise.Perlin3DNode]{Data: noise.Perlin3DNode{
		Time: nodes.ConstOutput[[]vector3.Float64]{Val: points},
		Seed: nodes.ConstOutput[int]{Val: seed},
	}}
	return nodes.GetNodeOutputPort[[]float64](&node, "Out").Value()
}

func TestPerlinSeedChangesTheNoiseRepeatably(t *testing.T) {
	assert.Equal(t, perlin3(1), perlin3(1))
	assert.NotEqual(t, perlin3(1), perlin3(2))
}

func TestPerlinFieldFollowsItsSeed(t *testing.T) {
	field := func(seed int) float64 {
		node := nodes.Struct[noise.Perlin3DFieldNode]{Data: noise.Perlin3DFieldNode{
			Seed: nodes.ConstOutput[int]{Val: seed},
		}}
		return nodes.GetNodeOutputPort[sample.Vec3ToFloat](&node, "Field").Value()(vector3.New(0.3, 0.7, 0.2))
	}
	assert.Equal(t, field(5), field(5))
	assert.NotEqual(t, field(5), field(6))
}
