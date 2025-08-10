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

func (d Subtract[T]) Out(out *nodes.StructOutput[vector3.Vector[T]]) {
	a := nodes.TryGetOutputValue(out, d.A, vector3.Zero[T]())
	b := nodes.TryGetOutputValue(out, d.B, vector3.Zero[T]())
	out.Set(a.Sub(b))
}

type SubtractToArrayNode[T vector.Number] struct {
	Amount nodes.Output[vector3.Vector[T]]
	Array  nodes.Output[[]vector3.Vector[T]]
}

func (cn SubtractToArrayNode[T]) Out(out *nodes.StructOutput[[]vector3.Vector[T]]) {
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
		total[i] = v.Sub(amount)
	}
	out.Set(total)
}
