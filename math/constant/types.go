package constant

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type ConstOutput[T any] struct {
	Ref             nodes.Node
	Val             T
	PortName        string
	PortDescription string
}

func (ConstOutput[T]) Version() int {
	return 0
}

func (co ConstOutput[T]) Value() T {
	return co.Val
}

func (co ConstOutput[T]) Node() nodes.Node {
	return co.Ref
}

func (co ConstOutput[T]) Name() string {
	return co.PortName
}

func (co ConstOutput[T]) Description() string {
	return co.PortDescription
}

func (so ConstOutput[T]) Type() string {
	return refutil.GetTypeWithPackage(new(T))
}

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[Vector3[float64]](factory)
	refutil.RegisterType[Vector3[int]](factory)
	refutil.RegisterType[Pi](factory)
	refutil.RegisterType[Quaternion](factory)

	generator.RegisterTypes(factory)
}
