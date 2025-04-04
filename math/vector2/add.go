package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type Sum = nodes.Struct[SumNodeData[float64]]

type SumNodeData[T vector.Number] struct {
	Values []nodes.Output[vector2.Vector[T]]
}

func (cn SumNodeData[T]) Out() nodes.StructOutput[vector2.Vector[T]] {
	var total vector2.Vector[T]
	for _, v := range cn.Values {
		total = total.Add(v.Value())
	}
	return nodes.NewStructOutput(total)
}

type ShiftArrayNode = nodes.Struct[ShiftArrayNodeData[float64]]

type ShiftArrayNodeData[T vector.Number] struct {
	Array  nodes.Output[[]vector2.Vector[T]]
	Amount nodes.Output[vector2.Vector[T]]
}

func (cn ShiftArrayNodeData[T]) Out() nodes.StructOutput[[]vector2.Vector[T]] {
	if cn.Array == nil {
		return nodes.NewStructOutput[[]vector2.Vector[T]](nil)
	}

	if cn.Amount == nil {
		return nodes.NewStructOutput(cn.Array.Value())
	}

	original := cn.Array.Value()
	amount := cn.Amount.Value()
	total := make([]vector2.Vector[T], len(original))
	for i, v := range original {
		total[i] = v.Add(amount)
	}
	return nodes.NewStructOutput(total)
}
