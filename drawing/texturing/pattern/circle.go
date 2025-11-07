package pattern

import (
	"fmt"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
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

type CircleNode[T any] struct {
	Dimensions nodes.Output[vector2.Int]
	Radius     nodes.Output[float64]

	Background nodes.Output[T]
	Fill       nodes.Output[T]

	InnerBorder          nodes.Output[T]
	InnerBorderThickness nodes.Output[float64]

	OuterBorder          nodes.Output[T]
	OuterBorderThickness nodes.Output[float64]
}

func (c CircleNode[T]) Texture(out *nodes.StructOutput[texturing.Texture[T]]) {
	dimensions := nodes.TryGetOutputValue(out, c.Dimensions, vector2.New(256, 256))
	if dimensions.MinComponent() <= 0 {
		out.CaptureError(texturing.InvalidDimension(dimensions))
		return
	}

	var t T
	circle := Circle[T]{
		Background:           nodes.TryGetOutputValue(out, c.Background, t),
		Radius:               nodes.TryGetOutputValue(out, c.Radius, 0.5),
		Fill:                 nodes.TryGetOutputValue(out, c.Fill, t),
		InnerBorder:          nodes.TryGetOutputValue(out, c.InnerBorder, t),
		InnerBorderThickness: nodes.TryGetOutputValue(out, c.InnerBorderThickness, 0),
		OuterBorder:          nodes.TryGetOutputValue(out, c.OuterBorder, t),
		OuterBorderThickness: nodes.TryGetOutputValue(out, c.OuterBorderThickness, 0),
	}

	out.Set(circle.Texture(dimensions))
}
