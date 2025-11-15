package coloring

import (
	"image/color"
	"math"
)

func ScaleRGB(c color.Color, amount float64) color.Color {
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

func ScaleColor(c color.Color, amount float64) color.Color {
	r, g, b, a := c.RGBA()

	rVal := math.Round(float64(r>>8) * amount)
	gVal := math.Round(float64(g>>8) * amount)
	bVal := math.Round(float64(b>>8) * amount)
	aVal := math.Round(float64(a>>8) * amount)

	return color.RGBA{
		R: uint8(rVal),
		G: uint8(gVal),
		B: uint8(bVal),
		A: uint8(aVal),
	}
}

func MultiplyRGBComponents(a, b color.Color) color.Color {
	rA, gA, bA, _ := a.RGBA()
	rB, gB, bB, _ := b.RGBA()

	rVal := math.Round(float64(rA>>8) * float64(rB>>8))
	gVal := math.Round(float64(gA>>8) * float64(gB>>8))
	bVal := math.Round(float64(bA>>8) * float64(bB>>8))

	return color.RGBA{
		R: uint8(rVal),
		G: uint8(gVal),
		B: uint8(bVal),
		A: 255,
	}
}

func SubtractColor(a, b color.Color) color.Color {
	rA, gA, bA, aA := a.RGBA()
	rB, gB, bB, aB := b.RGBA()

	rVal := (rA >> 8) - (rB >> 8)
	gVal := (gA >> 8) - (gB >> 8)
	bVal := (bA >> 8) - (bB >> 8)
	aVal := (aA >> 8) - (aB >> 8)

	return color.RGBA{
		R: uint8(rVal),
		G: uint8(gVal),
		B: uint8(bVal),
		A: uint8(aVal),
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

func Red(c color.Color) byte {
	r, _, _, _ := c.RGBA()
	return byte(r >> 8)
}

func Green(c color.Color) byte {
	_, g, _, _ := c.RGBA()
	return byte(g >> 8)
}

func Blue(c color.Color) byte {
	_, _, b, _ := c.RGBA()
	return byte(b >> 8)
}

func Alpha(c color.Color) byte {
	_, _, _, a := c.RGBA()
	return byte(a >> 8)
}

func RedEqual(c color.Color, val byte) bool {
	r, _, _, _ := c.RGBA()
	return byte(r>>8) == val
}

func RedGreaterThan(c color.Color, val byte) bool {
	r, _, _, _ := c.RGBA()
	return byte(r>>8) > val
}

func RedGreaterThanOrEqual(c color.Color, val byte) bool {
	r, _, _, _ := c.RGBA()
	return byte(r>>8) >= val
}

func RedLessThan(c color.Color, val byte) bool {
	r, _, _, _ := c.RGBA()
	return byte(r>>8) < val
}

func RedLessThanOrEqual(c color.Color, val byte) bool {
	r, _, _, _ := c.RGBA()
	return byte(r>>8) <= val
}

func GreenEqual(c color.Color, val byte) bool {
	_, g, _, _ := c.RGBA()
	return byte(g>>8) == val
}

func GreenGreaterThan(c color.Color, val byte) bool {
	_, g, _, _ := c.RGBA()
	return byte(g>>8) > val
}

func GreenGreaterThanOrEqual(c color.Color, val byte) bool {
	_, g, _, _ := c.RGBA()
	return byte(g>>8) >= val
}

func GreenLessThan(c color.Color, val byte) bool {
	_, g, _, _ := c.RGBA()
	return byte(g>>8) < val
}

func GreenLessThanOrEqual(c color.Color, val byte) bool {
	_, g, _, _ := c.RGBA()
	return byte(g>>8) <= val
}

func BlueEqual(c color.Color, val byte) bool {
	_, _, b, _ := c.RGBA()
	return byte(b>>8) == val
}

func BlueGreaterThan(c color.Color, val byte) bool {
	_, _, b, _ := c.RGBA()
	return byte(b>>8) > val
}

func BlueGreaterThanOrEqual(c color.Color, val byte) bool {
	_, _, b, _ := c.RGBA()
	return byte(b>>8) >= val
}

func BlueLessThan(c color.Color, val byte) bool {
	_, _, b, _ := c.RGBA()
	return byte(b>>8) < val
}

func BlueLessThanOrEqual(c color.Color, val byte) bool {
	_, _, b, _ := c.RGBA()
	return byte(b>>8) <= val
}

func AlphaEqual(c color.Color, val byte) bool {
	_, _, _, a := c.RGBA()
	return byte(a>>8) == val
}

func AlphaGreaterThan(c color.Color, val byte) bool {
	_, _, _, a := c.RGBA()
	return byte(a>>8) > val
}

func AlphaGreaterThanOrEqual(c color.Color, val byte) bool {
	_, _, _, a := c.RGBA()
	return byte(a>>8) >= val
}

func AlphaLessThan(c color.Color, val byte) bool {
	_, _, _, a := c.RGBA()
	return byte(a>>8) < val
}

func AlphaLessThanOrEqual(c color.Color, val byte) bool {
	_, _, _, a := c.RGBA()
	return byte(a>>8) <= val
}

func Interpolate(a, b color.Color, t float64) color.Color {
	rA, gA, bA, aA := a.RGBA()
	rB, gB, bB, aB := b.RGBA()

	rAF := float64(rA >> 8)
	gAF := float64(gA >> 8)
	bAF := float64(bA >> 8)
	aAF := float64(aA >> 8)

	rBF := float64(rB >> 8)
	gBF := float64(gB >> 8)
	bBF := float64(bB >> 8)
	aBF := float64(aB >> 8)

	rVal := math.Round(((rBF - rAF) * t) + rAF)
	gVal := math.Round(((gBF - gAF) * t) + gAF)
	bVal := math.Round(((bBF - bAF) * t) + bAF)
	aVal := math.Round(((aBF - aAF) * t) + aAF)

	return color.RGBA{
		R: uint8(rVal),
		G: uint8(gVal),
		B: uint8(bVal),
		A: uint8(aVal),
	}
}

func Greyscale(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	rF := float64(r >> 8)
	gF := float64(g >> 8)
	bF := float64(b >> 8)
	return (rF + gF + bF) / (255 * 3)
}
