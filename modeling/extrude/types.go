package extrude

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[ScrewNode]](factory)
	refutil.RegisterType[nodes.Struct[CircleNode]](factory)
	refutil.RegisterType[nodes.Struct[CircleAlongSplineNode]](factory)
	generator.RegisterTypes(factory)
}
