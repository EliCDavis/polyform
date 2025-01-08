package primitives

import "github.com/EliCDavis/polyform/refutil"

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[CubeNode](factory)
	refutil.RegisterType[HemisphereNode](factory)
	refutil.RegisterType[StripUVsNode](factory)
	return factory
}
