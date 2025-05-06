package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type SumNodeData[T vector.Number] struct {
	Values []nodes.Output[vector3.Vector[T]]
}

func (cn SumNodeData[T]) Out() nodes.StructOutput[vector3.Vector[T]] {
	var total vector3.Vector[T]
	for _, v := range cn.Values {
		total = total.Add(v.Value())
	}
	return nodes.NewStructOutput(total)
}

// ============================================================================

type AddToArrayNodeData[T vector.Number] struct {
	Amount nodes.Output[vector3.Vector[T]]
	Array  nodes.Output[[]vector3.Vector[T]]
}

func (cn AddToArrayNodeData[T]) Out() nodes.StructOutput[[]vector3.Vector[T]] {
	if cn.Array == nil {
		return nodes.NewStructOutput[[]vector3.Vector[T]](nil)
	}

	if cn.Amount == nil {
		return nodes.NewStructOutput(cn.Array.Value())
	}

	original := cn.Array.Value()
	amount := cn.Amount.Value()
	total := make([]vector3.Vector[T], len(original))
	for i, v := range original {
		total[i] = v.Add(amount)
	}
	return nodes.NewStructOutput(total)
}
