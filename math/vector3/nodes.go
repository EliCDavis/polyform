package vector3

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[New](factory)
	refutil.RegisterType[NewArray](factory)
	refutil.RegisterType[Sum](factory)
	refutil.RegisterType[ShiftArrayNode](factory)

	generator.RegisterTypes(factory)
}
