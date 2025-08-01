package vector2

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector2"
)

type SumNodeData[T vector.Number] struct {
	Values []nodes.Output[vector2.Vector[T]]
}

func (cn SumNodeData[T]) Out(out *nodes.StructOutput[vector2.Vector[T]]) {
	values := nodes.GetOutputValues(out, cn.Values)
	var total vector2.Vector[T]
	for _, v := range values {
		total = total.Add(v)
	}
	out.Set(total)
}

type AddToArrayNodeData[T vector.Number] struct {
	Array  nodes.Output[[]vector2.Vector[T]]
	Amount nodes.Output[vector2.Vector[T]]
}

func (cn AddToArrayNodeData[T]) Out(out *nodes.StructOutput[[]vector2.Vector[T]]) {
	if cn.Array == nil {
		return
	}

	original := nodes.GetOutputValue(out, cn.Array)
	if cn.Amount == nil {
		out.Set(original)
		return
	}

	amount := nodes.GetOutputValue(out, cn.Amount)
	total := make([]vector2.Vector[T], len(original))
	for i, v := range original {
		total[i] = v.Add(amount)
	}
	out.Set(total)
}
