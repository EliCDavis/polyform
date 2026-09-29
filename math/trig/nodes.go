package trig

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[SinNode]](factory)
	refutil.RegisterType[nodes.Struct[CosNode]](factory)
	refutil.RegisterType[nodes.Struct[TanNode]](factory)

	refutil.RegisterType[nodes.Struct[ArcSinNode]](factory)
	refutil.RegisterType[nodes.Struct[ArcCosNode]](factory)
	refutil.RegisterType[nodes.Struct[ArcTanNode]](factory)
	refutil.RegisterType[nodes.Struct[ArcTan2Node]](factory)

	refutil.RegisterType[nodes.Struct[DegreesToRadiansNode]](factory)
	refutil.RegisterType[nodes.Struct[RadiansToDegreesNode]](factory)

	generator.RegisterTypes(factory)
}
