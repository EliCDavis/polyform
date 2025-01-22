package vector

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type SumNode = nodes.Struct[vector3.Float64, SumData[float64]]

type SumData[T vector.Number] struct {
	Values []nodes.NodeOutput[vector3.Vector[T]]
}

func (cn SumData[T]) Process() (vector3.Vector[T], error) {
	var total vector3.Vector[T]
	for _, v := range cn.Values {
		total = total.Add(v.Value())
	}
	return total, nil
}

type ShiftArrayNode = nodes.Struct[[]vector3.Float64, ShiftArrayNodeData[float64]]

type ShiftArrayNodeData[T vector.Number] struct {
	Array  nodes.NodeOutput[[]vector3.Vector[T]]
	Amount nodes.NodeOutput[vector3.Vector[T]]
}

func (cn ShiftArrayNodeData[T]) Process() ([]vector3.Vector[T], error) {
	if cn.Array == nil {
		return nil, nil
	}

	if cn.Amount == nil {
		return cn.Array.Value(), nil
	}

	original := cn.Array.Value()
	amount := cn.Amount.Value()
	total := make([]vector3.Vector[T], len(original))
	for i, v := range original {
		total[i] = v.Add(amount)
	}
	return total, nil
}
