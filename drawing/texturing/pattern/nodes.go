package pattern

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[GridNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[GridNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[GridNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[GridNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[GridNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[RepeatNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[RepeatNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[RepeatNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[RepeatNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[RepeatNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[CircleNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[CircleNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[CircleNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[CircleNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[CircleNode[coloring.Color]]](factory)

	generator.RegisterTypes(factory)
}
