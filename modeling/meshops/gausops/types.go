package gausops

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[ColorGradingLutNode]](factory)
	refutil.RegisterType[nodes.Struct[ScaleNode]](factory)
	refutil.RegisterType[nodes.Struct[ScaleWithinRegionNode]](factory)
	refutil.RegisterType[nodes.Struct[RotateAttributeNode]](factory)
	refutil.RegisterType[nodes.Struct[FilterNode]](factory)

	generator.RegisterTypes(factory)
}
