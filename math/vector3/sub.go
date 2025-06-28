package vector3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Subtract[T vector.Number] struct {
	A nodes.Output[vector3.Vector[T]]
	B nodes.Output[vector3.Vector[T]]
}

func (d Subtract[T]) Out() nodes.StructOutput[vector3.Vector[T]] {
	a := nodes.TryGetOutputValue(d.A, vector3.Zero[T]())
	b := nodes.TryGetOutputValue(d.B, vector3.Zero[T]())
	return nodes.NewStructOutput(a.Sub(b))
}

type SubtractToArrayNodeData[T vector.Number] struct {
	Amount nodes.Output[vector3.Vector[T]]
	Array  nodes.Output[[]vector3.Vector[T]]
}

func (cn SubtractToArrayNodeData[T]) Out() nodes.StructOutput[[]vector3.Vector[T]] {
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
		total[i] = v.Sub(amount)
	}
	return nodes.NewStructOutput(total)
}
