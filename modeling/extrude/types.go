package extrude

import "github.com/EliCDavis/polyform/refutil"

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[ScrewNode](factory)
	refutil.RegisterType[CircleNode](factory)
	return factory
}
