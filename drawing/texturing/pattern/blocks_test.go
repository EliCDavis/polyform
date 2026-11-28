package pattern_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/drawing/texturing/pattern"
	"github.com/EliCDavis/polyform/math/norm"
	"github.com/EliCDavis/polyform/math/scalar"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func integer(v int) nodes.Output[int] {
	return nodes.ConstOutput[int]{Val: v}
}

func unorm(v float64) nodes.Output[norm.UFloat64] {
	return nodes.ConstOutput[norm.UFloat64]{Val: norm.NewUFloat64(v)}
}

func shiftedUV(w, h int, shift vector2.Float64) nodes.Output[uvTexture] {
	return nodes.ConstOutput[uvTexture]{Val: texturing.Map(texturing.UVs(w, h), func(uv vector2.Float64) vector2.Float64 {
		return uv.Add(shift)
	})}
}

func distinctInRow(tex texturing.Texture[float64], y int) int {
	seen := map[float64]bool{}
	for x := range tex.Width() {
		seen[tex.Get(x, y)] = true
	}
	return len(seen)
}

func blocks(uv nodes.Output[uvTexture]) pattern.BlocksNode {
	return pattern.BlocksNode{
		UV:                 uv,
		Rows:               integer(4),
		Units:              integer(8),
		JointChance:        unorm(0.5),
		MaxUnits:           integer(3),
		RowHeightVariation: unorm(0.6),
		RowSlide:           unorm(1),
		Seed:               integer(3),
	}
}

func TestBlocksTileAcrossWholeUVPeriods(t *testing.T) {
	base := blocks(plainUV(64, 32))
	shifted := blocks(shiftedUV(64, 32, vector2.New(1., -2.)))

	for _, port := range []string{"Random", "Edge Distance"} {
		a := texture[float64](t, base, port)
		b := texture[float64](t, shifted, port)
		for y := range 32 {
			for x := range 64 {
				assert.InDelta(t, a.Get(x, y), b.Get(x, y), 1e-9, "%s at %d,%d", port, x, y)
			}
		}
	}
}

func TestBlocksWithNoJointsSpanTheRow(t *testing.T) {
	node := blocks(plainUV(64, 32))
	node.JointChance = unorm(0)
	node.MaxUnits = integer(8)
	node.RowHeightVariation = unorm(0)

	random := texture[float64](t, node, "Random")
	for _, y := range []int{2, 10, 18, 26} {
		assert.Equal(t, 1, distinctInRow(random, y))
	}
}

func TestBlocksJointAtEveryUnit(t *testing.T) {
	node := blocks(plainUV(64, 32))
	node.JointChance = unorm(1)
	node.RowHeightVariation = unorm(0)

	assert.Equal(t, 8, distinctInRow(texture[float64](t, node, "Random"), 2))
}

func TestBlocksNeverExceedMaxUnits(t *testing.T) {
	node := blocks(plainUV(64, 32))
	node.JointChance = unorm(0)
	node.MaxUnits = integer(2)
	node.RowHeightVariation = unorm(0)

	assert.Equal(t, 4, distinctInRow(texture[float64](t, node, "Random"), 2))
}

func TestBlocksEdgeDistanceIsInPixels(t *testing.T) {
	node := blocks(plainUV(8, 8))
	node.Rows = integer(1)
	node.Units = integer(4)
	node.JointChance = unorm(1)
	node.RowHeightVariation = unorm(0)
	node.RowSlide = unorm(0)

	distance := texture[float64](t, node, "Edge Distance")
	assert.InDelta(t, 0.5, distance.Get(0, 4), 1e-9)
	assert.InDelta(t, 0.5, distance.Get(1, 4), 1e-9)
	assert.InDelta(t, 0.5, distance.Get(4, 0), 1e-9)
}

func TestBlocksEdgeDistanceStaysInPixelsUnderTransformedUVs(t *testing.T) {
	transformed := func(scale vector2.Float64, degrees float64) pattern.BlocksNode {
		node := blocks(nodes.GetNodeOutputPort[uvTexture](&nodes.Struct[pattern.TransformUVNode]{Data: pattern.TransformUVNode{
			In:       plainUV(64, 32),
			Scale:    nodes.ConstOutput[vector2.Float64]{Val: scale},
			Rotation: number(degrees),
		}}, "UV"))
		node.Rows = integer(1)
		node.Units = integer(1)
		node.RowHeightVariation = unorm(0)
		node.RowSlide = unorm(0)
		return node
	}

	scaled := texture[float64](t, transformed(vector2.New(2., 1.), 0), "Edge Distance")
	assert.InDelta(t, 0.5, scaled.Get(16, 16), 1e-9)
	assert.InDelta(t, 4.5, scaled.Get(20, 16), 1e-9)

	turned := texture[float64](t, transformed(vector2.New(1., 1.), 90), "Edge Distance")
	assert.InDelta(t, 0.5, turned.Get(0, 16), 1e-9)
	assert.InDelta(t, 5.5, turned.Get(5, 16), 1e-9)
	assert.InDelta(t, 5.5, turned.Get(32, 5), 1e-9)
}

func TestBlocksEdgeDistanceIgnoresWhereUVsWrap(t *testing.T) {
	wrapped := nodes.ConstOutput[uvTexture]{Val: texturing.Map(texturing.UVs(64, 64), func(uv vector2.Float64) vector2.Float64 {
		return vector2.New(scalar.Fract(uv.X()*4), scalar.Fract(uv.Y()*4))
	})}
	node := blocks(wrapped)
	node.Rows = integer(1)
	node.Units = integer(1)
	node.RowHeightVariation = unorm(0)
	node.RowSlide = unorm(0)

	distance := texture[float64](t, node, "Edge Distance")
	assert.InDelta(t, 0.5, distance.Get(0, 8), 1e-9)
	assert.InDelta(t, 3.5, distance.Get(19, 8), 1e-9)
}

func TestBlocksRoundPullsCornersIn(t *testing.T) {
	node := blocks(plainUV(16, 16))
	node.Rows = integer(1)
	node.Units = integer(1)
	node.RowHeightVariation = unorm(0)
	node.RowSlide = unorm(0)

	sharp := texture[float64](t, node, "Edge Distance").Get(0, 0)
	node.Round = number(4)
	rounded := texture[float64](t, node, "Edge Distance").Get(0, 0)

	assert.InDelta(t, 0.5, sharp, 1e-9)
	assert.Less(t, rounded, 0.)
}

func TestBlocksLocalUVSpansEachBlock(t *testing.T) {
	node := blocks(plainUV(8, 8))
	node.Rows = integer(1)
	node.Units = integer(2)
	node.JointChance = unorm(1)
	node.RowHeightVariation = unorm(0)
	node.RowSlide = unorm(0)

	local := texture[vector2.Float64](t, node, "Local UV")
	assert.InDelta(t, 0.125, local.Get(0, 0).X(), 1e-9)
	assert.InDelta(t, 0.125, local.Get(4, 0).X(), 1e-9)
	assert.InDelta(t, 0.0625, local.Get(0, 0).Y(), 1e-9)
}

func TestRehashMatchesEqualInputsAndChangesWithSeed(t *testing.T) {
	in := nodes.ConstOutput[texturing.Texture[float64]]{Val: texturing.FromArray([]float64{0.25, 0.75, 0.25, 0.75}, 4, 1)}

	a := texture[float64](t, pattern.RehashNode{Texture: in, Seed: integer(1)}, "Value")
	b := texture[float64](t, pattern.RehashNode{Texture: in, Seed: integer(2)}, "Value")

	assert.Equal(t, a.Get(0, 0), a.Get(2, 0))
	assert.NotEqual(t, a.Get(0, 0), a.Get(1, 0))
	assert.NotEqual(t, a.Get(0, 0), b.Get(0, 0))
}
