package vector4

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector4"
)

type SumNode[T vector.Number] struct {
	Values []nodes.Output[vector4.Vector[T]]
}

func (cn SumNode[T]) Out(out *nodes.StructOutput[vector4.Vector[T]]) {
	values := nodes.GetOutputValues(out, cn.Values)
	var total vector4.Vector[T]
	for _, v := range values {
		total = total.Add(v)
	}
	out.Set(total)
}

type AddToArrayNode[T vector.Number] struct {
	Array  nodes.Output[[]vector4.Vector[T]]
	Amount nodes.Output[vector4.Vector[T]]
}

func (cn AddToArrayNode[T]) Out(out *nodes.StructOutput[[]vector4.Vector[T]]) {
	if cn.Array == nil {
		return
	}

	original := nodes.GetOutputValue(out, cn.Array)
	if cn.Amount == nil {
		out.Set(original)
		return
	}

	amount := nodes.GetOutputValue(out, cn.Amount)
	total := make([]vector4.Vector[T], len(original))
	for i, v := range original {
		total[i] = v.Add(amount)
	}
	out.Set(total)
}
