package texturing

import (
	"cmp"
	"image"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector1"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func boxBlur(src image.Image, dst *image.RGBA) {
	ConvolveImage(src, func(x, y int, vals []color.Color) {
		x1y1 := coloring.ScaleRGB(vals[0], 1./9)
		x2y1 := coloring.ScaleRGB(vals[1], 1./9)
		x3y1 := coloring.ScaleRGB(vals[2], 1./9)

		x1y2 := coloring.ScaleRGB(vals[3], 1./9)
		x2y2 := coloring.ScaleRGB(vals[4], 1./9)
		x3y2 := coloring.ScaleRGB(vals[5], 1./9)

		x1y3 := coloring.ScaleRGB(vals[6], 1./9)
		x2y3 := coloring.ScaleRGB(vals[7], 1./9)
		x3y3 := coloring.ScaleRGB(vals[8], 1./9)

		dst.Set(x, y, coloring.AddRGB(
			x1y1, x2y1, x3y1,
			x1y2, x2y2, x3y2,
			x1y3, x2y3, x3y3,
		))
	})
}

func BoxBlur(src image.Image) image.Image {
	dst := image.NewRGBA(src.Bounds())
	boxBlur(src, dst)
	return dst
}

func BoxBlurNTimes(src image.Image, iterations int) image.Image {
	if iterations < 1 {
		return src
	}
	dst := image.NewRGBA(src.Bounds())
	boxBlur(src, dst)
	if iterations == 1 {
		return dst
	}

	dst2 := image.NewRGBA(src.Bounds())
	for i := 1; i < iterations; i++ {
		if i%2 == 0 {
			boxBlur(dst2, dst)
		} else {
			boxBlur(dst, dst2)
		}
	}
	if iterations%2 == 0 {
		return dst2
	}
	return dst
}

// Generates a normalized 1D Gaussian kernel
func gaussianKernel(radius int, sigma float64) []float64 {
	kernel := make([]float64, 2*radius+1)
	var sum float64

	sigma2 := sigma * sigma * 2
	for i := -radius; i <= radius; i++ {
		v := math.Exp(-(float64(i * i)) / sigma2)
		kernel[i+radius] = v
		sum += v
	}
	for i := range kernel {
		kernel[i] /= sum
	}
	return kernel
}

func clamp[T cmp.Ordered](i, minimum, maximum T) T {
	return max(min(i, maximum), minimum)
}

// Applies a 1D convolution along x or y
func convolve1DGaussian[T any](space vector.Space[T], src Texture[T], dst Texture[T], kernel []float64, horizontal bool) {
	radius := len(kernel) / 2

	if horizontal {
		src.ScanParallel(func(x, y int, v T) {
			var accum T
			for k := -radius; k <= radius; k++ {
				sx := clamp(x+k, 0, src.width-1)
				weighted := space.Scale(src.Get(sx, y), kernel[k+radius])
				accum = space.Add(accum, weighted)
			}
			dst.Set(x, y, accum)
		})
	} else {
		src.ScanParallel(func(x, y int, v T) {
			var accum T
			for k := -radius; k <= radius; k++ {
				sy := clamp(y+k, 0, src.height-1)
				weighted := space.Scale(src.Get(x, sy), kernel[k+radius])
				accum = space.Add(accum, weighted)
			}
			dst.Set(x, y, accum)
		})
	}
}

// GaussianBlur applies a Gaussian blur to the texture
func RadialGaussianBlur[T any](space vector.Space[T], src Texture[T], radius int, sigma float64) Texture[T] {
	kernel := gaussianKernel(radius, sigma)

	tmp := Empty[T](src.width, src.height)
	out := Empty[T](src.width, src.height)

	// Horizontal then vertical
	convolve1DGaussian(space, src, tmp, kernel, true)
	convolve1DGaussian(space, tmp, out, kernel, false)

	return out
}

type GaussianBlurFloatNode struct {
	Texture nodes.Output[Texture[float64]] `description:"Texture to blur"`
	Radius  nodes.Output[int]              `description:"Size of the kernel in pixels"`
	Signma  nodes.Output[float64]          `description:"standard deviation of the gaussian"`
}

func (n GaussianBlurFloatNode) Value(out *nodes.StructOutput[Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	out.Set(RadialGaussianBlur(
		vector1.Space[float64]{},
		nodes.GetOutputValue(out, n.Texture),
		nodes.TryGetOutputValue(out, n.Radius, 1),
		nodes.TryGetOutputValue(out, n.Signma, 1),
	))
}

type GaussianBlurFloat2Node struct {
	Texture nodes.Output[Texture[vector2.Float64]] `description:"Texture to blur"`
	Radius  nodes.Output[int]                      `description:"Size of the kernel in pixels"`
	Signma  nodes.Output[float64]                  `description:"standard deviation of the gaussian"`
}

func (n GaussianBlurFloat2Node) Value(out *nodes.StructOutput[Texture[vector2.Float64]]) {
	if n.Texture == nil {
		return
	}
	out.Set(RadialGaussianBlur(
		vector2.Space[float64]{},
		nodes.GetOutputValue(out, n.Texture),
		nodes.TryGetOutputValue(out, n.Radius, 1),
		nodes.TryGetOutputValue(out, n.Signma, 1),
	))
}

type GaussianBlurFloat3Node struct {
	Texture nodes.Output[Texture[vector3.Float64]] `description:"Texture to blur"`
	Radius  nodes.Output[int]                      `description:"Size of the kernel in pixels"`
	Signma  nodes.Output[float64]                  `description:"standard deviation of the gaussian"`
}

func (n GaussianBlurFloat3Node) Value(out *nodes.StructOutput[Texture[vector3.Float64]]) {
	if n.Texture == nil {
		return
	}
	out.Set(RadialGaussianBlur(
		vector3.Space[float64]{},
		nodes.GetOutputValue(out, n.Texture),
		nodes.TryGetOutputValue(out, n.Radius, 1),
		nodes.TryGetOutputValue(out, n.Signma, 1),
	))
}

type GaussianBlurColorNode struct {
	Texture nodes.Output[Texture[coloring.Color]] `description:"Texture to blur"`
	Radius  nodes.Output[int]                     `description:"Size of the kernel in pixels"`
	Signma  nodes.Output[float64]                 `description:"standard deviation of the gaussian"`
}

func (n GaussianBlurColorNode) Value(out *nodes.StructOutput[Texture[coloring.Color]]) {
	if n.Texture == nil {
		return
	}
	out.Set(RadialGaussianBlur(
		coloring.Space{},
		nodes.GetOutputValue(out, n.Texture),
		nodes.TryGetOutputValue(out, n.Radius, 1),
		nodes.TryGetOutputValue(out, n.Signma, 1),
	))
}
