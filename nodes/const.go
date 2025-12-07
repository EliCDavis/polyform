package nodes

import "github.com/EliCDavis/polyform/refutil"

type ConstOutput[T any] struct {
	Ref             Node
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

func (co ConstOutput[T]) Node() Node {
	return co.Ref
}

func (co ConstOutput[T]) Name() string {
	return co.PortName
}

func (co ConstOutput[T]) Description() string {
	return co.PortDescription
}

func (so ConstOutput[T]) Type() string {
	resolver := refutil.TypeResolution{
		IncludePackage: true,
		IncludePointer: false,
	}
	return resolver.Resolve(new(T))
}
