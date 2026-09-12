package aabb

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[NewNode]](factory)
	refutil.RegisterType[nodes.Struct[FromMinMaxNode]](factory)
	refutil.RegisterType[nodes.Struct[FromPointsNode]](factory)
	refutil.RegisterType[nodes.Struct[ExpandNode]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode]](factory)

	generator.RegisterTypes(factory)
}
