package texturing_test

import (
	"testing"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/stretchr/testify/assert"
)

func floats(values ...float64) nodes.Output[texturing.Texture[float64]] {
	return nodes.ConstOutput[texturing.Texture[float64]]{Val: texturing.FromArray(values, len(values), 1)}
}

func number(v float64) nodes.Output[float64] {
	return nodes.ConstOutput[float64]{Val: v}
}

func result[T, N any](data N) (texturing.Texture[T], []string) {
	out := nodes.GetNodeOutputPort[texturing.Texture[T]](&nodes.Struct[N]{Data: data}, "Result")
	value := out.Value()
	var errs []string
	if observed, ok := out.(nodes.ObservableExecution); ok {
		errs = observed.ExecutionReport().Errors
	}
	return value, errs
}

func row(tex texturing.Texture[float64]) []float64 {
	values := make([]float64, tex.Width())
	for x := range values {
		values[x] = tex.Get(x, 0)
	}
	return values
}

func TestUVNodeHoldsPixelCentres(t *testing.T) {
	uv := nodes.GetNodeOutputPort[texturing.Texture[vector2.Float64]](&nodes.Struct[texturing.UVNode]{Data: texturing.UVNode{
		Dimensions: nodes.ConstOutput[vector2.Int]{Val: vector2.New(2, 4)},
	}}, "UV").Value()

	assert.Equal(t, vector2.New(0.25, 0.125), uv.Get(0, 0))
	assert.Equal(t, vector2.New(0.75, 0.875), uv.Get(1, 3))
}

func TestSelectFloat2SeparatesComponents(t *testing.T) {
	uv := texturing.FromArray([]vector2.Float64{vector2.New(0.25, 0.5), vector2.New(0.75, 0.5)}, 2, 1)
	node := &nodes.Struct[texturing.SelectFloat2Node]{Data: texturing.SelectFloat2Node{
		Texture: nodes.ConstOutput[texturing.Texture[vector2.Float64]]{Val: uv},
	}}

	x := nodes.GetNodeOutputPort[texturing.Texture[float64]](node, "X").Value()
	y := nodes.GetNodeOutputPort[texturing.Texture[float64]](node, "Y").Value()

	assert.Equal(t, 0.25, x.Get(0, 0))
	assert.Equal(t, 0.75, x.Get(1, 0))
	assert.Equal(t, 0.5, y.Get(0, 0))
	assert.Equal(t, 0.5, y.Get(1, 0))
	assert.Equal(t, 2, nodes.GetNodeOutputPort[int](node, "Width").Value())
}

func TestSubtractRefusesMismatchedSizes(t *testing.T) {
	_, errs := result[float64](texturing.SubtractFloat1Node{A: floats(1, 2), B: floats(1)})

	assert.NotEmpty(t, errs)
}

func TestTextureMathNodes(t *testing.T) {
	sub, _ := result[float64](texturing.SubtractFloat1Node{A: floats(3, 2), B: floats(1, 5)})
	assert.Equal(t, []float64{2, -3}, row(sub))

	lo, _ := result[float64](texturing.MinFloat1Node{Textures: []nodes.Output[texturing.Texture[float64]]{floats(3, 2), floats(1, 5)}})
	assert.Equal(t, []float64{1, 2}, row(lo))

	hi, _ := result[float64](texturing.MaxFloat1Node{Textures: []nodes.Output[texturing.Texture[float64]]{floats(3, 2), floats(1, 5)}})
	assert.Equal(t, []float64{3, 5}, row(hi))

	remapped, _ := result[float64](texturing.RemapFloat1Node{Texture: floats(0, 0.5, 2), InMin: number(0), InMax: number(1), OutMin: number(10), OutMax: number(20), Clamp: nodes.ConstOutput[bool]{Val: true}})
	assert.Equal(t, []float64{10, 15, 20}, row(remapped))

	clamped, _ := result[float64](texturing.ClampFloat1Node{Texture: floats(-1, 0.5, 2), Min: number(0), Max: number(1)})
	assert.Equal(t, []float64{0, 0.5, 1}, row(clamped))

	powered, _ := result[float64](texturing.PowerFloat1Node{Texture: floats(0.5, 3), Exponent: number(2)})
	assert.Equal(t, []float64{0.25, 9}, row(powered))

	stepped, _ := result[float64](texturing.SmoothstepFloat1Node{Texture: floats(0, 0.5, 1), Low: number(0.25), High: number(0.75)})
	assert.Equal(t, []float64{0, 0.5, 1}, row(stepped))
}

func TestMixUsesTheFactorTextureOverTheAmount(t *testing.T) {
	even, _ := result[float64](texturing.MixFloat1Node{A: floats(0, 0), B: floats(10, 10), Amount: number(0.5)})
	assert.Equal(t, []float64{5, 5}, row(even))

	masked, _ := result[float64](texturing.MixFloat1Node{A: floats(0, 0), B: floats(10, 10), Factor: floats(0, 1), Amount: number(0.5)})
	assert.Equal(t, []float64{0, 10}, row(masked))
}

func TestMixColorBlendsPerPixel(t *testing.T) {
	colors := func(c coloring.Color) nodes.Output[texturing.Texture[coloring.Color]] {
		return nodes.ConstOutput[texturing.Texture[coloring.Color]]{Val: texturing.FromArray([]coloring.Color{c, c}, 2, 1)}
	}

	mixed, _ := result[coloring.Color](texturing.MixColorNode{
		A:      colors(coloring.Color{R: 1, A: 1}),
		B:      colors(coloring.Color{B: 1, A: 1}),
		Factor: floats(0, 1),
	})

	assert.Equal(t, coloring.Color{R: 1, A: 1}, mixed.Get(0, 0))
	assert.Equal(t, coloring.Color{B: 1, A: 1}, mixed.Get(1, 0))
}

func TestScaleColorKeepsAlpha(t *testing.T) {
	colors := nodes.ConstOutput[texturing.Texture[coloring.Color]]{Val: texturing.FromArray([]coloring.Color{{R: 1, G: 0.5, B: 0.2, A: 1}}, 1, 1)}

	scaled := nodes.GetNodeOutputPort[texturing.Texture[coloring.Color]](&nodes.Struct[texturing.ScaleColorNode]{Data: texturing.ScaleColorNode{
		Texture: colors,
		Scale:   floats(0.5),
	}}, "Result").Value()
	uniform := nodes.GetNodeOutputPort[texturing.Texture[coloring.Color]](&nodes.Struct[texturing.ScaleColorUniformNode]{Data: texturing.ScaleColorUniformNode{
		Texture: colors,
		Scale:   number(0.5),
	}}, "Result").Value()

	assert.Equal(t, coloring.Color{R: 0.5, G: 0.25, B: 0.1, A: 1}, scaled.Get(0, 0))
	assert.Equal(t, coloring.Color{R: 0.5, G: 0.25, B: 0.1, A: 1}, uniform.Get(0, 0))
}
