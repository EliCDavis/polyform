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

	refutil.RegisterType[nodes.Struct[Half[int]]](factory)
	refutil.RegisterType[nodes.Struct[Half[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Double[int]]](factory)
	refutil.RegisterType[nodes.Struct[Double[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Scale[int]]](factory)
	refutil.RegisterType[nodes.Struct[Scale[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Dot]](factory)

	refutil.RegisterType[nodes.Struct[Length[int]]](factory)
	refutil.RegisterType[nodes.Struct[Length[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Distance[float64]]](factory)
	refutil.RegisterType[nodes.Struct[Distance[int]]](factory)
	refutil.RegisterType[nodes.Struct[Distances[float64]]](factory)
	refutil.RegisterType[nodes.Struct[Distances[int]]](factory)
	refutil.RegisterType[nodes.Struct[DistancesToArray[float64]]](factory)
	refutil.RegisterType[nodes.Struct[DistancesToArray[int]]](factory)
	refutil.RegisterType[nodes.Struct[DistancesToNodes[float64]]](factory)
	refutil.RegisterType[nodes.Struct[DistancesToNodes[int]]](factory)

	refutil.RegisterType[nodes.Struct[Inverse[int]]](factory)
	refutil.RegisterType[nodes.Struct[Inverse[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Subtract[int]]](factory)
	refutil.RegisterType[nodes.Struct[Subtract[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SubtractToArrayNode[int]]](factory)
	refutil.RegisterType[nodes.Struct[SubtractToArrayNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[Normalize]](factory)
	refutil.RegisterType[nodes.Struct[NormalizeArray]](factory)

	generator.RegisterTypes(factory)
}
