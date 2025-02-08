package trig

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[SinArray](factory)
	refutil.RegisterType[CosArray](factory)

	generator.RegisterTypes(factory)
}
