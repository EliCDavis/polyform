package nodes

import (
	"github.com/EliCDavis/vector"
)

type Sum[T vector.Number] struct {
	StructData[T]

	Values []NodeOutput[T]
}

func (cn *Sum[T]) Out() NodeOutput[T] {
	return &StructNodeOutput[T]{Definition: cn}
}

func (cn Sum[T]) Process() (T, error) {
	var total T
	for _, v := range cn.Values {
		total += v.Data()
	}
	return total, nil
}

// ============================================================================
type Difference[T vector.Number] struct {
	StructData[T]

	A NodeOutput[T]
	B NodeOutput[T]
}

func (cn *Difference[T]) Out() NodeOutput[T] {
	return &StructNodeOutput[T]{Definition: cn}
}

func (cn Difference[T]) Process() (T, error) {
	return cn.A.Data() - cn.B.Data(), nil
}

type Divide[T vector.Number] struct {
	StructData[T]

	Dividend NodeOutput[T]
	Divisor  NodeOutput[T]
}

func (cn *Divide[T]) Out() NodeOutput[T] {
	return &StructNodeOutput[T]{Definition: cn}
}

func (cn Divide[T]) Process() (T, error) {
	return cn.Dividend.Data() / cn.Divisor.Data(), nil
}
