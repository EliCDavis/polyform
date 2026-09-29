package math

import (
	"errors"
	gomath "math"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

var cantDivideByZeroErr = errors.New("can't divide by 0")

type DivideNode[T vector.Number] struct {
	Dividend nodes.LiftedPort[T] `description:"the number being divided"`
	Divisor  nodes.LiftedPort[T] `description:"number doing the dividing"`
}

func (DivideNode[T]) Description() string {
	return "Dividend / Divisor"
}

func (cn DivideNode[T]) Float(out *nodes.Lifted[float64]) {
	dividedByZero := false
	nodes.Zip2(out, cn.Dividend, cn.Divisor, func(a, b T) float64 {
		if b == 0 {
			dividedByZero = true
			return 0
		}
		return float64(a / b)
	})
	if dividedByZero {
		out.CaptureError(cantDivideByZeroErr)
	}
}

func (cn DivideNode[T]) Int(out *nodes.Lifted[int]) {
	dividedByZero := false
	nodes.Zip2(out, cn.Dividend, cn.Divisor, func(a, b T) int {
		if b == 0 {
			dividedByZero = true
			return 0
		}
		return int(gomath.Round(float64(a / b)))
	})
	if dividedByZero {
		out.CaptureError(cantDivideByZeroErr)
	}
}
