package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type SumNode[T vector.Number] struct {
	Values []nodes.LiftedPort[vector4.Vector[T]] `description:"The vectors to add. Arrays are added element by element."`
}

func (cn SumNode[T]) Description() string {
	return "Adds vectors together."
}

func (cn SumNode[T]) Out(out *nodes.Lifted[vector4.Vector[T]]) {
	nodes.ZipAll(out, cn.Values, func(values []vector4.Vector[T]) vector4.Vector[T] {
		var total vector4.Vector[T]
		for _, v := range values {
			total = total.Add(v)
		}
		return total
	})
}
