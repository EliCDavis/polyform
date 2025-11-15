package texturing

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func texture(
	out *nodes.StructOutput[Texture[float64]],
	noiseFunc func(coord, size vector2.Float64, folds, octaves int, persistence, offset, seed float64) float64,
	Width nodes.Output[int],
	Height nodes.Output[int],
	Scale nodes.Output[vector2.Float64],
	Folds nodes.Output[int],
	Octaves nodes.Output[int],
	Persistence nodes.Output[float64],
	Offset nodes.Output[float64],
	Seed nodes.Output[float64],
	Polar nodes.Output[bool],
) {
	width := nodes.TryGetOutputValue(out, Width, 1)
	height := nodes.TryGetOutputValue(out, Height, 1)

	if width <= 0 {
		out.CaptureError(fmt.Errorf("invalid width dimension: %d", width))
		return
	}

	if height <= 0 {
		out.CaptureError(fmt.Errorf("invalid height dimension: %d", height))
		return
	}
	t := Empty[float64](width, height)

	scale := nodes.TryGetOutputValue(out, Scale, vector2.One[float64]())
	offset := nodes.TryGetOutputValue(out, Offset, 0)
	seed := nodes.TryGetOutputValue(out, Seed, 0)
	persistance := nodes.TryGetOutputValue(out, Persistence, 0.5)
	octaves := nodes.TryGetOutputValue(out, Octaves, 3)
	folds := nodes.TryGetOutputValue(out, Folds, 0)
	dimensions := vector2.New(width, height).ToFloat64()
	polar := nodes.TryGetOutputValue(out, Polar, false)

	if polar {
		center := dimensions.Scale(0.5)
		diagonal := center.Length()
		t.MutateParallel(func(x, y int, v float64) float64 {
			p := vector2.New(x, y).ToFloat64().Sub(center)
			cord := vector2.New(p.Length()/diagonal, math.Abs(math.Atan2(p.Y(), p.X())))
			return noiseFunc(
				cord,
				scale,
				folds,
				octaves,
				persistance,
				offset,
				seed,
			)
		})
	} else {
		t.MutateParallel(func(x, y int, v float64) float64 {
			return noiseFunc(
				vector2.New(x, y).ToFloat64().DivByVector(dimensions),
				scale,
				folds,
				octaves,
				persistance,
				offset,
				seed,
			)
		})
	}

	out.Set(t)
}

type NoiseNode struct {
	Width       nodes.Output[int]             `description:"number of pixels in the x direction of the texture"`
	Height      nodes.Output[int]             `description:"number of pixels in the y direction of the texture"`
	Scale       nodes.Output[vector2.Float64] `description:"the X and Y scale of the first octave noise"`
	Folds       nodes.Output[int]             `description:"the number of folds (offsetting the noise negatively and taking the absolute value)"`
	Octaves     nodes.Output[int]             `description:"the number of iterations"`
	Persistence nodes.Output[float64]         `description:"the strength of each subsequent iteration"`
	Offset      nodes.Output[float64]         `description:"the offset of the points, can be used to animate the noise"`
	Seed        nodes.Output[float64]         `description:"RNG seed"`
	Polar       nodes.Output[bool]            `description:"Whether or not to use polar coordinates to sample the random function"`
}

func (n NoiseNode) Value(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Value,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) ValueDescription() string {
	return "A texture generated as a sum of Value noise functions with increasing frequencies and decreasing amplitudes"
}

func (n NoiseNode) Perlin(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Perlin,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Simplex(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Simplex,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Cellular(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Cellular,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Cellular2(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Cellular2,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Cellular3(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Cellular3,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Cellular4(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Cellular4,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Cellular5(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Cellular5,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Cellular6(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Cellular6,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}

func (n NoiseNode) Voronoise(out *nodes.StructOutput[Texture[float64]]) {
	texture(
		out,
		noise.Voronoise,
		n.Width, n.Height, n.Scale, n.Folds, n.Octaves, n.Persistence, n.Offset, n.Seed, n.Polar)
}
