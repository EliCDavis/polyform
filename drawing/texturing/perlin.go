package texturing

import (
	"math"

	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type SeamlessPerlinNode struct {
	Dimensions nodes.Output[int]     `default:"256"`
	Positive   nodes.Output[float64] `default:"1"`
	Negative   nodes.Output[float64]
	Octaves    nodes.Output[int]     `default:"3"`
	Frequency  nodes.Output[float64] `description:"Cycles per pixel, rounded to whole cycles across the image." default:"0.015625"`
	Seed       nodes.Output[int]
}

func (an SeamlessPerlinNode) Description() string {
	return "Perlin noise texture that tiles without a visible seam."
}

func (an SeamlessPerlinNode) Out(out *nodes.StructOutput[Texture[float64]]) {
	dim := nodes.GetOutputValue(out, an.Dimensions)
	n := noise.NewTilingNoiseWithSeed(
		dim,
		nodes.GetOutputValue(out, an.Frequency),
		nodes.GetOutputValue(out, an.Octaves),
		nodes.GetOutputValue(out, an.Seed),
	)

	tex := Empty[float64](dim, dim)
	negative := nodes.GetOutputValue(out, an.Negative)
	positive := nodes.GetOutputValue(out, an.Positive)
	valueRange := positive - negative

	tex.MutateParallel(func(x, y int, _ float64) float64 {
		p := (n.Noise(x, y) * 0.5) + 0.5
		return negative + (valueRange * p)
	})
	out.Set(tex)
}

// ============================================================================

type PerlinNode struct {
	Width     nodes.Output[int]     `default:"1"`
	Height    nodes.Output[int]     `default:"1"`
	Positive  nodes.Output[float64] `default:"1"`
	Negative  nodes.Output[float64]
	Frequency nodes.Output[vector2.Float64] `description:"Cycles per pixel." default:"0.015625,0.015625"`
}

func (n PerlinNode) Description() string {
	return "Perlin noise texture."
}

func (n PerlinNode) Out(out *nodes.StructOutput[Texture[float64]]) {
	tex := Empty[float64](
		nodes.GetOutputValue(out, n.Width),
		nodes.GetOutputValue(out, n.Height),
	)

	frequncy := nodes.GetOutputValue(out, n.Frequency)
	negative := nodes.GetOutputValue(out, n.Negative)
	positive := nodes.GetOutputValue(out, n.Positive)
	valueRange := positive - negative

	for y := range tex.height {
		for x := range tex.width {
			v := vector2.New(x, y).ToFloat64().MultByVector(frequncy)
			// Perlin2D is centred on 0.5 and reaches √2/2 either side of it.
			p := (noise.Perlin2D(v)-0.5)/math.Sqrt2 + 0.5
			tex.Set(x, y, negative+(valueRange*p))
		}
	}

	out.Set(tex)
}
