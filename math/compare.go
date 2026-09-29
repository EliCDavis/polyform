package math

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector"
)

type CompareNode[T vector.Number] struct {
	A nodes.LiftedPort[T]
	B nodes.LiftedPort[T]
}

func (CompareNode[T]) Description() string {
	return "Compares two numbers, one bool output per relation."
}

func (CompareNode[T]) Keywords() []string {
	return []string{"condition", "if"}
}

func (n CompareNode[T]) Equal(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b T) bool { return a == b })
}

func (n CompareNode[T]) NotEqual(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b T) bool { return a != b })
}

func (n CompareNode[T]) Greater(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b T) bool { return a > b })
}

func (n CompareNode[T]) GreaterOrEqual(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b T) bool { return a >= b })
}

func (n CompareNode[T]) Less(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b T) bool { return a < b })
}

func (n CompareNode[T]) LessOrEqual(out *nodes.Lifted[bool]) {
	nodes.Zip2(out, n.A, n.B, func(a, b T) bool { return a <= b })
}

type BoolToNumberNode struct {
	In nodes.LiftedPort[bool]
}

func (BoolToNumberNode) Description() string {
	return "True as 1, false as 0."
}

func (n BoolToNumberNode) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, n.In, func(v bool) int {
		if v {
			return 1
		}
		return 0
	})
}

func (n BoolToNumberNode) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, n.In, func(v bool) float64 {
		if v {
			return 1
		}
		return 0
	})
}

type NotNode struct {
	In nodes.LiftedPort[bool]
}

func (NotNode) Description() string {
	return "Inverts a bool."
}

func (n NotNode) Out(out *nodes.Lifted[bool]) {
	nodes.Zip1(out, n.In, func(v bool) bool { return !v })
}
