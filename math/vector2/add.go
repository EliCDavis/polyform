package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type SumNode[T vector.Number] struct {
	Values []nodes.LiftedPort[vector2.Vector[T]] `description:"The vectors to add. Arrays are added element by element."`
}

func (cn SumNode[T]) Description() string {
	return "Adds vectors together."
}

func (cn SumNode[T]) Out(out *nodes.Lifted[vector2.Vector[T]]) {
	nodes.ZipAll(out, cn.Values, func(values []vector2.Vector[T]) vector2.Vector[T] {
		var total vector2.Vector[T]
		for _, v := range values {
			total = total.Add(v)
		}
		return total
	})
}
