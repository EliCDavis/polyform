package texturing

import (
	"image/color"
	"math"
)

func MultiplyRGBByConstant(c color.Color, amount float64) color.Color {
	r, g, b, a := c.RGBA()

	rVal := math.Round(float64(r>>8) * amount)
	gVal := math.Round(float64(g>>8) * amount)
	bVal := math.Round(float64(b>>8) * amount)

	return color.RGBA{
		R: uint8(rVal),
		G: uint8(gVal),
		B: uint8(bVal),
		A: uint8(a >> 8),
	}
}

func AddRGB(colors ...color.Color) color.Color {
	var rVal uint8 = 0
	var gVal uint8 = 0
	var bVal uint8 = 0

	for _, c := range colors {
		r, g, b, _ := c.RGBA()

		rVal += uint8(r >> 8)
		gVal += uint8(g >> 8)
		bVal += uint8(b >> 8)
	}

	return color.RGBA{
		R: uint8(rVal),
		G: uint8(gVal),
		B: uint8(bVal),
		A: uint8(255),
	}
}
