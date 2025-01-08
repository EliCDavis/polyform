package repeat

import "github.com/EliCDavis/polyform/refutil"

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[MeshNode](factory)
	refutil.RegisterType[CircleNode](factory)
	return factory
}
