package vecn3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type SumNode = nodes.StructNode[vector3.Float64, SumData[float64]]

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
