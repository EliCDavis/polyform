package texturing

import (
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type SeamlessPerlinNode struct {
	Dimensions nodes.Output[int]
	Positive   nodes.Output[float64]
	Negative   nodes.Output[float64]
	Octaves    nodes.Output[int]
	Frequency  nodes.Output[float64]
}

func (an SeamlessPerlinNode) Out(out *nodes.StructOutput[Texture[float64]]) {
	dim := nodes.TryGetOutputValue(out, an.Dimensions, 256)
	n := noise.NewTilingNoise(
		dim,
		nodes.TryGetOutputValue(out, an.Frequency, 1/64.),
		nodes.TryGetOutputValue(out, an.Octaves, 3),
	)

	tex := Empty[float64](dim, dim)
	negative := nodes.TryGetOutputValue(out, an.Negative, 0)
	positive := nodes.TryGetOutputValue(out, an.Positive, 1)
	valueRange := positive - negative

	for y := range dim {
		for x := range dim {
			p := (n.Noise(x, y) * 0.5) + 0.5

			tex.Set(x, y, negative+(valueRange*p))
		}
	}
	out.Set(tex)
}

// ============================================================================

type PerlinNode struct {
	Width     nodes.Output[int]
	Height    nodes.Output[int]
	Positive  nodes.Output[float64]
	Negative  nodes.Output[float64]
	Frequency nodes.Output[vector2.Float64]
}

func (n PerlinNode) Out(out *nodes.StructOutput[Texture[float64]]) {
	tex := Empty[float64](
		nodes.TryGetOutputValue(out, n.Width, 1),
		nodes.TryGetOutputValue(out, n.Height, 1),
	)

	frequncy := nodes.TryGetOutputValue(out, n.Frequency, vector2.New(1., 1.))
	negative := nodes.TryGetOutputValue(out, n.Negative, 0)
	positive := nodes.TryGetOutputValue(out, n.Positive, 1)
	valueRange := positive - negative

	for y := range tex.height {
		for x := range tex.width {
			v := vector2.New(x, y).ToFloat64().MultByVector(frequncy)
			p := (noise.Perlin2D(v) * 0.5) + 0.5
			// log.Println(p)
			tex.Set(x, y, negative+(valueRange*p))
		}
	}

	out.Set(tex)
}
