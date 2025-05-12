package trig

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[SinArray](factory)
	refutil.RegisterType[CosArray](factory)
	refutil.RegisterType[nodes.Struct[TanArray]](factory)

	refutil.RegisterType[nodes.Struct[ArcSinArray]](factory)
	refutil.RegisterType[nodes.Struct[ArcCosArray]](factory)
	refutil.RegisterType[nodes.Struct[ArcTanArray]](factory)

	generator.RegisterTypes(factory)
}
