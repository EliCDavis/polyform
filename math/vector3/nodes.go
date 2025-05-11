package vector3

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[NewNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[ArrayFromComponentsNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[ArrayFromComponentsNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[SumNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SumNodeData[int]]](factory)

	refutil.RegisterType[nodes.Struct[AddToArrayNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[AddToArrayNodeData[float64]]](factory)

	refutil.RegisterType[nodes.Struct[ArrayFromNodesNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[ArrayFromNodesNodeData[float64]]](factory)

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

	generator.RegisterTypes(factory)
}
