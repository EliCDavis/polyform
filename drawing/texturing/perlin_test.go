package texturing_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func perlinRange(node texturing.PerlinNode) (low, high float64) {
	tex := nodes.GetNodeOutputPort[texturing.Texture[float64]](&nodes.Struct[texturing.PerlinNode]{Data: node}, "Out").Value()
	low, high = tex.Get(0, 0), tex.Get(0, 0)
	for y := range tex.Height() {
		for x := range tex.Width() {
			low, high = min(low, tex.Get(x, y)), max(high, tex.Get(x, y))
		}
	}
	return low, high
}

func TestPerlinNodeVariesAtItsDefaultFrequency(t *testing.T) {
	low, high := perlinRange(texturing.PerlinNode{
		Width:  nodes.ConstOutput[int]{Val: 256},
		Height: nodes.ConstOutput[int]{Val: 256},
	})

	assert.Greater(t, high-low, 0.3)
}

func TestPerlinNodeStaysBetweenNegativeAndPositive(t *testing.T) {
	low, high := perlinRange(texturing.PerlinNode{
		Width:     nodes.ConstOutput[int]{Val: 512},
		Height:    nodes.ConstOutput[int]{Val: 512},
		Frequency: nodes.ConstOutput[vector2.Float64]{Val: vector2.New(0.0371, 0.0433)},
		Negative:  nodes.ConstOutput[float64]{Val: 2},
		Positive:  nodes.ConstOutput[float64]{Val: 6},
	})

	assert.GreaterOrEqual(t, low, 2.)
	assert.LessOrEqual(t, high, 6.)
	assert.Less(t, low, 3.)
	assert.Greater(t, high, 5.)
}
