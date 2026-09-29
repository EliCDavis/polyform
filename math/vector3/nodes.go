package vector3

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[NewNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[SumNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SumNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[Select[int]]](factory)
	refutil.RegisterType[nodes.Struct[Select[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Half[int]]](factory)
	refutil.RegisterType[nodes.Struct[Half[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Double[int]]](factory)
	refutil.RegisterType[nodes.Struct[Double[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Dot]](factory)

	refutil.RegisterType[nodes.Struct[Length[int]]](factory)
	refutil.RegisterType[nodes.Struct[Length[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Distance[float64]]](factory)
	refutil.RegisterType[nodes.Struct[Distance[int]]](factory)

	refutil.RegisterType[nodes.Struct[Inverse[int]]](factory)
	refutil.RegisterType[nodes.Struct[Inverse[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Subtract[int]]](factory)
	refutil.RegisterType[nodes.Struct[Subtract[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Scale[int]]](factory)
	refutil.RegisterType[nodes.Struct[Scale[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Normalize]](factory)
	refutil.RegisterType[nodes.Struct[NormalizeArray]](factory)

	refutil.RegisterType[nodes.Struct[Multiply[int]]](factory)
	refutil.RegisterType[nodes.Struct[Multiply[float64]]](factory)

	generator.RegisterTypes(factory)
}
