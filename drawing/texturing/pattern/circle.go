package pattern

import (
	"fmt"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/vector/vector2"
)

type Circle[T any] struct {
	Radius float64
	Fill   T

	InnerBorder          T
	InnerBorderThickness float64

	OuterBorder          T
	OuterBorderThickness float64

	Background T
}

func (c Circle[T]) Texture(dimensions vector2.Int) texturing.Texture[T] {
	if dimensions.MinComponent() < 0 {
		panic(fmt.Errorf("can not create circle: %w", texturing.InvalidDimension(dimensions)))
	}

	tex := texturing.Empty[T](dimensions.X(), dimensions.Y())
	if dimensions.MinComponent() == 0 {
		return tex
	}

	ranges := []struct {
		max float64
		val T
	}{}

	maxRange := 0.
	if c.Radius > 0 {
		maxRange += c.Radius
		ranges = append(ranges, struct {
			max float64
			val T
		}{
			max: maxRange,
			val: c.Fill,
		})
	}

	if c.InnerBorderThickness > 0 {
		maxRange += c.InnerBorderThickness
		ranges = append(ranges, struct {
			max float64
			val T
		}{
			max: maxRange,
			val: c.InnerBorder,
		})
	}

	if c.OuterBorderThickness > 0 {
		maxRange += c.OuterBorderThickness
		ranges = append(ranges, struct {
			max float64
			val T
		}{
			max: maxRange,
			val: c.OuterBorder,
		})
	}

	center := dimensions.ToFloat64().Scale(0.5)
	tex.MutateParallel(func(x, y int, v T) T {
		pix := vector2.New(x, y)

		dist := pix.ToFloat64().Distance(center)
		p := (dist * 2) / float64(dimensions.MinComponent())

		for _, r := range ranges {
			if p < r.max {
				return r.val
			}
		}
		return c.Background
	})

	return tex
}
