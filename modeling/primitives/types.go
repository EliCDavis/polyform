package primitives

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[CubeNode](factory)
	refutil.RegisterType[HemisphereNode](factory)
	refutil.RegisterType[StripUVsNode](factory)
	generator.RegisterTypes(factory)
}
