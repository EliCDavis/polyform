package pattern

import (
	"fmt"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/vector/vector2"
)

func Grid[T any](element texturing.Texture[T], repeat vector2.Int, targetDimensions vector2.Int) texturing.Texture[T] {
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
		percent := vector2.New(x, y).
			ToFloat64().
			DivByVector(targetElementDimensions)
		percent = percent.Sub(percent.Floor())

		pix := percent.MultByVector(elementDimension).RoundToInt()

		return element.Get(pix.X(), pix.Y())
	})

	return tex
}
