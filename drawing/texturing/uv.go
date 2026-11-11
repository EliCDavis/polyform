package texturing

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func Map[T, R any](in Texture[T], f func(T) R) Texture[R] {
	out := Empty[R](in.width, in.height)
	out.MutateParallel(func(x, y int, _ R) R {
		return f(in.Get(x, y))
	})
	return out
}

// UVs puts each pixel's centre in [0, 1], with V growing down the image as glTF's does.
func UVs(width, height int) Texture[vector2.Float64] {
	uv := Empty[vector2.Float64](width, height)
	uv.MutateParallel(func(x, y int, _ vector2.Float64) vector2.Float64 {
		return vector2.New((float64(x)+0.5)/float64(width), (float64(y)+0.5)/float64(height))
	})
	return uv
}

type UVNode struct {
	Dimensions nodes.Output[vector2.Int] `default:"256,256"`
}

func (n UVNode) Description() string {
	return "A texture whose every pixel holds its own UV coordinate, 0 to 1 across and down. The starting point for procedural patterns and UV warps."
}

func (n UVNode) UV(out *nodes.StructOutput[Texture[vector2.Float64]]) {
	dimensions := nodes.GetOutputValue(out, n.Dimensions)
	if dimensions.MinComponent() <= 0 {
		out.CaptureError(InvalidDimension(dimensions))
		return
	}
	out.Set(UVs(dimensions.X(), dimensions.Y()))
}
