package curves

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[CatmullRomSplineNode](factory)
	refutil.RegisterType[LengthNode](factory)

	generator.RegisterTypes(factory)
}
