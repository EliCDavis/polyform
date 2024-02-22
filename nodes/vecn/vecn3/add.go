package vecn3

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
)

type Sum[T vector.Number] struct {
	nodes.StructData[vector3.Vector[T]]

	Values []nodes.NodeOutput[vector3.Vector[T]]
}

func (cn *Sum[T]) Out() nodes.NodeOutput[vector3.Vector[T]] {
	return &nodes.StructNodeOutput[vector3.Vector[T]]{Definition: cn}
}

func (cn Sum[T]) Process() (vector3.Vector[T], error) {
	var total vector3.Vector[T]
	for _, v := range cn.Values {
		total = total.Add(v.Data())
	}
	return total, nil
}
