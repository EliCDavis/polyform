package pattern

import (
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type RectanglesNode[T any] struct {
	Positions nodes.Output[[]vector2.Float64]
	Value     nodes.Output[T]
	Size      nodes.Output[vector2.Float64]
	In        nodes.Output[texturing.Texture[T]]
}

func (node RectanglesNode[T]) Out(out *nodes.StructOutput[texturing.Texture[T]]) {
	if node.In == nil {
		out.CaptureError(nodes.UnsetInputError{
			Input: node.In,
		})
		return
	}

	inputTex := nodes.GetOutputValue(out, node.In)
	imgSize := vector2.New(inputTex.Width()-1, inputTex.Height()-1).ToFloat64()
	size := nodes.TryGetOutputValue(out, node.Size, vector2.New(0.5, 0.5)).Scale(0.5)

	var val T
	if node.Value != nil {
		val = nodes.GetOutputValue(out, node.Value)
	}

	outputTex := inputTex.Copy()
	positions := nodes.GetOutputValue(out, node.Positions)
	for _, pos := range positions {
		start := pos.
			Sub(size).
			Clamp(0, 1).
			MultByVector(imgSize).
			RoundToInt()

		end := pos.
			Add(size).
			Clamp(0, 1).
			MultByVector(imgSize).
			RoundToInt()

		for y := start.Y(); y < end.Y(); y++ {
			for x := start.X(); x < end.X(); x++ {
				outputTex.Set(x, y, val)
			}
		}
	}

	out.Set(outputTex)
}
