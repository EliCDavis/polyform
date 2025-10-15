package texturing

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
)

func DivideColor(a, b Texture[coloring.Color]) (*Texture[coloring.Color], error) {
	if !resolutionsMatch([]Texture[coloring.Color]{a, b}) {
		return nil, ErrMismatchDimensions
	}

	result := Empty[coloring.Color](a.width, a.height)
	a.ScanParallel(func(x, y int, v coloring.Color) {
		other := b.Get(x, y)
		result.Set(x, y, coloring.Color{
			R: v.R / other.R,
			G: v.G / other.G,
			B: v.B / other.B,
			A: v.A / other.A,
		})
	})

	return &result, nil
}

func MaxColor(a Texture[coloring.Color], other float64) Texture[coloring.Color] {

	result := Empty[coloring.Color](a.width, a.height)
	a.ScanParallel(func(x, y int, v coloring.Color) {
		result.Set(x, y, coloring.Color{
			R: max(v.R, other),
			G: max(v.G, other),
			B: max(v.B, other),
			A: max(v.A, other),
		})
	})

	return result
}

func ClampColor(a Texture[coloring.Color], minimum, maximum float64) Texture[coloring.Color] {

	result := Empty[coloring.Color](a.width, a.height)
	a.ScanParallel(func(x, y int, v coloring.Color) {
		result.Set(x, y, coloring.Color{
			R: min(max(v.R, minimum), maximum),
			G: min(max(v.G, minimum), maximum),
			B: min(max(v.B, minimum), maximum),
			A: min(max(v.A, minimum), maximum),
		})
	})

	return result
}
