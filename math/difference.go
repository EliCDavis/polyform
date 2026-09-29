package math

import (
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type SubtractNode[T vector.Number] struct {
	A nodes.LiftedPort[T] `description:"The value being subtracted from."`
	B nodes.LiftedPort[T] `description:"The value being subtracted."`
}

func (cn SubtractNode[T]) Description() string {
	return "A - B"
}

func (an SubtractNode[T]) Float(out *nodes.Lifted[float64]) {
	nodes.Zip2(out, an.A, an.B, func(a, b T) float64 {
		return float64(a - b)
	})
}

func (an SubtractNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip2(out, an.A, an.B, func(a, b T) int {
		return int(gomath.Round(float64(a - b)))
	})
}
