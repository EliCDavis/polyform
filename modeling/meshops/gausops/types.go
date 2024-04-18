package gausops

import "github.com/EliCDavis/polyform/refutil"

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ColorGradingLutNode](factory)
	refutil.RegisterType[ScaleNode](factory)

	return factory
}
