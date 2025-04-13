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
	refutil.RegisterType[nodes.Struct[NewArrayNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[NewArrayNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[SumNodeData[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SumNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[ShiftArrayNodeData[int]]](factory)
	refutil.RegisterType[nodes.Struct[ShiftArrayNodeData[float64]]](factory)

	generator.RegisterTypes(factory)
}
