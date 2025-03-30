package constant

import (
	"math"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type ConstOutput[T any] struct {
	Ref      nodes.Node
	Val      T
	PortName string
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
func (so ConstOutput[T]) Type() string {
	return refutil.GetTypeWithPackage(new(T))
}

type Pi struct{}

func (Pi) Inputs() map[string]nodes.InputPort {
	return nil
}

func (p *Pi) Outputs() map[string]nodes.OutputPort {
	return map[string]nodes.OutputPort{
		"Pi": ConstOutput[float64]{
			Ref:      p,
			Val:      math.Pi,
			PortName: "Pi",
		},

		"Pi / 2": ConstOutput[float64]{
			Ref:      p,
			Val:      math.Pi / 2,
			PortName: "Pi / 2",
		},

		"2Pi": ConstOutput[float64]{
			Ref:      p,
			Val:      math.Pi * 2,
			PortName: "2Pi",
		},
	}
}

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[Pi](factory)

	// factory.RegisterBuilder("github.com/EliCDavis/polyform/generator/constant.Py", )

	// refutil.RegisterTypeWithBuilder(factory, func() Pi {
	// 	return Pi{Value: math.Pi}
	// })

	generator.RegisterTypes(factory)
}
