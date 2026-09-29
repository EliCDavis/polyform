package vector4

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

	refutil.RegisterType[nodes.Struct[Subtract[int]]](factory)
	refutil.RegisterType[nodes.Struct[Subtract[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Scale[int]]](factory)
	refutil.RegisterType[nodes.Struct[Scale[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Normalize]](factory)
	refutil.RegisterType[nodes.Struct[NormalizeArray]](factory)

	generator.RegisterTypes(factory)
}
