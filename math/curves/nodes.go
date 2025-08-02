package curves

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[CatmullRomSplineNode]](factory)
	refutil.RegisterType[nodes.Struct[LengthNode]](factory)

	generator.RegisterTypes(factory)
}
