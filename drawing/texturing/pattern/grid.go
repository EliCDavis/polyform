package pattern

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type Grid[T any] struct {
	HorizontalLines     int
	VerticalLines       int
	BackgroundValue     T
	LineValue           T
	HorizontalLineWidth float64
	VerticalLineWidth   float64
}

func (c Grid[T]) Texture(dimensions vector2.Int) texturing.Texture[T] {
	if dimensions.MinComponent() < 0 {
		panic(fmt.Errorf("can not create grid: %w", texturing.InvalidDimension(dimensions)))
	}

	tex := texturing.Empty[T](dimensions.X(), dimensions.Y())
	if dimensions.MinComponent() == 0 {
		return tex
	}

	// There's no lines to render, so we'll
	if (c.HorizontalLineWidth <= 0 || c.HorizontalLines < 1) && (c.VerticalLineWidth <= 0 || c.VerticalLines < 1) {
		tex.MutateParallel(func(x, y int, v T) T {
			return c.BackgroundValue
		})
		return tex
	}

	// One of the lines will fill the entire texture
	if (c.HorizontalLineWidth >= 1 && c.HorizontalLines > 0) || (c.VerticalLineWidth >= 1 && c.VerticalLines > 0) {
		tex.MutateParallel(func(x, y int, v T) T {
			return c.LineValue
		})
		return tex
	}

	fDim := dimensions.ToFloat64()
	lineThicknesses := fDim.MultByVector(vector2.New(c.HorizontalLineWidth, c.VerticalLineWidth)).Scale(0.5)
	lineSpacing := fDim.DivByVector(vector2.New(float64(c.HorizontalLines+1), float64(c.VerticalLines+1)))

	tex.MutateParallel(func(x, y int, v T) T {
		pix := vector2.New(x, y).ToFloat64()
		closestLine := pix.
			DivByVector(lineSpacing).
			Round()

		closestLine = closestLine.SetX(math.Max(math.Min(closestLine.X(), float64(c.HorizontalLines)), 1))
		closestLine = closestLine.SetY(math.Max(math.Min(closestLine.Y(), float64(c.VerticalLines)), 1))

		dist := closestLine.
			MultByVector(lineSpacing).
			Sub(pix).
			Abs()

		onLine := (c.HorizontalLines > 0 && dist.X() <= lineThicknesses.X()) || (c.VerticalLines > 0 && dist.Y() <= lineThicknesses.Y())
		if onLine {
			return c.LineValue
		}
		return c.BackgroundValue
	})

	return tex
}

type GridNode[T any] struct {
	HorizontalLines     nodes.Output[int]         `default:"5"`
	VerticalLines       nodes.Output[int]         `default:"5"`
	Dimensions          nodes.Output[vector2.Int] `default:"256,256"`
	BackgroundValue     nodes.Output[T]
	LineValue           nodes.Output[T]
	HorizontalLineWidth nodes.Output[float64] `default:"0.05"`
	VerticalLineWidth   nodes.Output[float64] `default:"0.05"`
}

func (gnd GridNode[T]) Texture(out *nodes.StructOutput[texturing.Texture[T]]) {
	dimensions := nodes.GetOutputValue(out, gnd.Dimensions)
	if dimensions.MinComponent() <= 0 {
		out.CaptureError(texturing.InvalidDimension(dimensions))
		return
	}

	var t T
	grid := Grid[T]{
		HorizontalLines:     nodes.GetOutputValue(out, gnd.HorizontalLines),
		VerticalLines:       nodes.GetOutputValue(out, gnd.VerticalLines),
		BackgroundValue:     nodes.TryGetOutputValue(out, gnd.BackgroundValue, t),
		LineValue:           nodes.TryGetOutputValue(out, gnd.LineValue, t),
		HorizontalLineWidth: nodes.GetOutputValue(out, gnd.HorizontalLineWidth),
		VerticalLineWidth:   nodes.GetOutputValue(out, gnd.VerticalLineWidth),
	}

	out.Set(grid.Texture(dimensions))
}
