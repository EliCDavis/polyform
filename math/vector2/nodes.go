package vector2

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[NewNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[ArrayFromComponentsNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[ArrayFromComponentsNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[SumNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SumNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[AddToArrayNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[AddToArrayNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[ArrayFromNodesNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[ArrayFromNodesNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Select[int]]](factory)
	refutil.RegisterType[nodes.Struct[Select[float64]]](factory)

	refutil.RegisterType[nodes.Struct[SelectArray[int]]](factory)
	refutil.RegisterType[nodes.Struct[SelectArray[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Subtract[int]]](factory)
	refutil.RegisterType[nodes.Struct[Subtract[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SubtractToArrayNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[SubtractToArrayNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Scale[int]]](factory)
	refutil.RegisterType[nodes.Struct[Scale[float64]]](factory)

	generator.RegisterTypes(factory)
}
