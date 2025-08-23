package coloring

import (
	"image/color"
	"math"
)

/*
type Space[T any] interface {
	Distance(a, b T) float64
	Add(a, b T) T
	Sub(a, b T) T
	Scale(a T, amount float64) T
	Dot(a, b T) float64
	Length(a T) float64
	Normalized(a T) T
	Lerp(a, b T, time float64) T
}

*/

type Space struct {
}

func (c Space) Distance(a, b color.Color) float64 {
	return c.Length(c.Sub(b, a))
}

func (Space) Add(a, b color.Color) color.Color {
	return AddRGB(a, b)
}

func (Space) Sub(a, b color.Color) color.Color {
	return SubtractColor(a, b)
}

func (Space) Scale(a color.Color, amount float64) color.Color {
	return MultiplyColorByConstant(a, amount)
}

func (Space) Dot(a, b color.Color) float64 {
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

	return (rAF * rBF) + (gAF * gBF) + (bAF * bBF) + (aAF * aBF)
}

func (Space) Length(c color.Color) float64 {
	r, g, b, a := c.RGBA()

	rF := float64(r >> 8)
	gF := float64(g >> 8)
	bF := float64(b >> 8)
	aF := float64(a >> 8)

	return math.Sqrt((rF * rF) + (gF * gF) + (bF * bF) + (aF * aF))
}

func (Space) Normalized(a color.Color) color.Color {
	// Shit. uh. shit
	// I'm not sure what we can do about this given we're not in floating point
	// land.

	// At the moment. I don't know when I'd used `ColorSpace` struct in vector
	// math that would require normalizaiton. If it ever comes time, we probably
	// will just change this function to do whatever it needs done.

	return a
}

func (Space) Lerp(a, b color.Color, time float64) color.Color {
	return Interpolate(a, b, time)
}
