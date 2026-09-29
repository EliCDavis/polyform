package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type SumNode[T vector.Number] struct {
	Values []nodes.LiftedPort[vector3.Vector[T]] `description:"The vectors to add. Arrays are added element by element."`
}

func (cn SumNode[T]) Description() string {
	return "Adds vectors together."
}

func (cn SumNode[T]) Out(out *nodes.Lifted[vector3.Vector[T]]) {
	nodes.ZipAll(out, cn.Values, func(values []vector3.Vector[T]) vector3.Vector[T] {
		var total vector3.Vector[T]
		for _, v := range values {
			total = total.Add(v)
		}
		return total
	})
}
