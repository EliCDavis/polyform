package pattern

import (
	"fmt"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func Repeat[T any](element texturing.Texture[T], repeat vector2.Int, targetDimensions vector2.Int) texturing.Texture[T] {
	if targetDimensions.MinComponent() <= 0 {
		panic(fmt.Errorf("can not create grid: %w", texturing.InvalidDimension(targetDimensions)))
	}

	if repeat.MinComponent() <= 0 {
		panic(fmt.Errorf("can not repeat texture element on a grid %dx%d times", repeat.X(), repeat.Y()))
	}

	elementDimension := vector2.New(element.Width(), element.Height()).ToFloat64()
	targetElementDimensions := targetDimensions.
		ToFloat64().
		DivByVector(repeat.ToFloat64())

	tex := texturing.Empty[T](targetDimensions.X(), targetDimensions.Y())
	tex.MutateParallel(func(x, y int, v T) T {
		pix := vector2.New(x, y).
			ToFloat64().
			DivByVector(targetElementDimensions).
			Fract().
			MultByVector(elementDimension).
			FloorToInt()

		return element.Get(pix.X(), pix.Y())
	})

	return tex
}

type RepeatNode[T any] struct {
	Dimensions nodes.Output[vector2.Int]
	Repeat     nodes.Output[vector2.Int]
	Element    nodes.Output[texturing.Texture[T]]
}

func (gnd RepeatNode[T]) Texture(out *nodes.StructOutput[texturing.Texture[T]]) {
	if gnd.Element == nil {
		return
	}

	dimensions := nodes.TryGetOutputValue(out, gnd.Dimensions, vector2.New(256, 256))
	if dimensions.MinComponent() <= 0 {
		out.CaptureError(texturing.InvalidDimension(dimensions))
		return
	}

	repeat := nodes.TryGetOutputValue(out, gnd.Repeat, vector2.New(2, 2))
	if repeat.MinComponent() <= 0 {
		out.CaptureError(fmt.Errorf("can't repeat something %d times", repeat.MinComponent()))
		return
	}

	out.Set(Repeat(
		nodes.GetOutputValue(out, gnd.Element),
		repeat,
		dimensions,
	))
}
