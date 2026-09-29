package math

import (
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type MultiplyNode[T vector.Number] struct {
	Values []nodes.LiftedPort[T] `description:"The values to multiply together, in order. Arrays are multiplied element by element."`
}

func (cn MultiplyNode[T]) Description() string {
	return "Multiplies two or more values together."
}

func (cn MultiplyNode[T]) product(vals []T) T {
	if len(vals) == 0 {
		return 0
	}

	total := vals[0]
	for i := 1; i < len(vals); i++ {
		total *= vals[i]
	}
	return total
}

func (cn MultiplyNode[T]) Float(out *nodes.Lifted[float64]) {
	nodes.ZipAll(out, cn.Values, func(vals []T) float64 {
		return float64(cn.product(vals))
	})
}

func (cn MultiplyNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.ZipAll(out, cn.Values, func(vals []T) int {
		return int(gomath.Round(float64(cn.product(vals))))
	})
}
