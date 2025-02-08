package experimental

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ShiftNode](factory)
	refutil.RegisterType[BrushedMetalNode](factory)
	refutil.RegisterType[SampleNode](factory)
	refutil.RegisterType[SeamlessPerlinNode](factory)

	generator.RegisterTypes(factory)
}
