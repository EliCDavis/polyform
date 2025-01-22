package vector

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[SumNode](factory)
	refutil.RegisterType[New](factory)
	refutil.RegisterType[ShiftArrayNode](factory)

	generator.RegisterTypes(factory)
}
