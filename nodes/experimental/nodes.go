package experimental

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ShiftNode]](factory)
	refutil.RegisterType[nodes.Struct[BrushedMetalNode]](factory)
	refutil.RegisterType[nodes.Struct[SampleNode]](factory)
	refutil.RegisterType[nodes.Struct[SeamlessPerlinNode]](factory)

	generator.RegisterTypes(factory)
}
