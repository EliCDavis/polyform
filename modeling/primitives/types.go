package primitives

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[CubeNode](factory)
	refutil.RegisterType[QuadNode](factory)
	refutil.RegisterType[StripUVsNode](factory)

	refutil.RegisterType[CylinderNode](factory)
	refutil.RegisterType[HemisphereNode](factory)
	refutil.RegisterType[UvSphereNode](factory)

	refutil.RegisterType[CircleNode](factory)
	refutil.RegisterType[CircleUVsNode](factory)

	generator.RegisterTypes(factory)
}
