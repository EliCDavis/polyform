package texturing

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
)

func GaussianBlur(src image.Image) image.Image {
	dst := image.NewRGBA(src.Bounds())
	Convolve(src, func(x, y int, vals []color.Color) {
		x1y1 := coloring.MultiplyRGBByConstant(vals[0], 1./16)
		x2y1 := coloring.MultiplyRGBByConstant(vals[1], 2./16)
		x3y1 := coloring.MultiplyRGBByConstant(vals[2], 1./16)

		x1y2 := coloring.MultiplyRGBByConstant(vals[3], 2./16)
		x2y2 := coloring.MultiplyRGBByConstant(vals[4], 4./16)
		x3y2 := coloring.MultiplyRGBByConstant(vals[5], 2./16)

		x1y3 := coloring.MultiplyRGBByConstant(vals[6], 1./16)
		x2y3 := coloring.MultiplyRGBByConstant(vals[7], 2./16)
		x3y3 := coloring.MultiplyRGBByConstant(vals[8], 1./16)

		dst.Set(x, y, coloring.AddRGB(
			x1y1, x2y1, x3y1,
			x1y2, x2y2, x3y2,
			x1y3, x2y3, x3y3,
		))
	})
	return dst
}

func boxBlur(src image.Image, dst *image.RGBA) {
	Convolve(src, func(x, y int, vals []color.Color) {
		x1y1 := coloring.MultiplyRGBByConstant(vals[0], 1./9)
		x2y1 := coloring.MultiplyRGBByConstant(vals[1], 1./9)
		x3y1 := coloring.MultiplyRGBByConstant(vals[2], 1./9)

		x1y2 := coloring.MultiplyRGBByConstant(vals[3], 1./9)
		x2y2 := coloring.MultiplyRGBByConstant(vals[4], 1./9)
		x3y2 := coloring.MultiplyRGBByConstant(vals[5], 1./9)

		x1y3 := coloring.MultiplyRGBByConstant(vals[6], 1./9)
		x2y3 := coloring.MultiplyRGBByConstant(vals[7], 1./9)
		x3y3 := coloring.MultiplyRGBByConstant(vals[8], 1./9)

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
