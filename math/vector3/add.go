package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type SumNode[T vector.Number] struct {
	Values []nodes.Output[vector3.Vector[T]]
}

func (cn SumNode[T]) Out(out *nodes.StructOutput[vector3.Vector[T]]) {
	values := nodes.GetOutputValues(out, cn.Values)
	var total vector3.Vector[T]
	for _, v := range values {
		total = total.Add(v)
	}
	out.Set(total)
}

// ============================================================================

type AddToArrayNode[T vector.Number] struct {
	Amount nodes.Output[vector3.Vector[T]]
	Array  nodes.Output[[]vector3.Vector[T]]
}

func (cn AddToArrayNode[T]) Out(out *nodes.StructOutput[[]vector3.Vector[T]]) {
	if cn.Array == nil {
		return
	}

	original := nodes.GetOutputValue(out, cn.Array)
	if cn.Amount == nil {
		out.Set(original)
		return
	}

	amount := nodes.GetOutputValue(out, cn.Amount)
	total := make([]vector3.Vector[T], len(original))
	for i, v := range original {
		total[i] = v.Add(amount)
	}
	out.Set(total)
}
