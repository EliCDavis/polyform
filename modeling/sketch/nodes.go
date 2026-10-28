package sketch

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[RoundedRectangleNode]](factory)
	refutil.RegisterType[nodes.Struct[TransformNode]](factory)
	refutil.RegisterType[nodes.Struct[StrokeNode]](factory)
	refutil.RegisterType[nodes.Struct[BevelNode]](factory)
	refutil.RegisterType[nodes.Struct[BevelPathNode]](factory)
	refutil.RegisterType[nodes.Struct[FillNode]](factory)
	refutil.RegisterType[nodes.Struct[UnionNode]](factory)
	refutil.RegisterType[nodes.Struct[IntersectionNode]](factory)
	refutil.RegisterType[nodes.Struct[SubtractNode]](factory)
	refutil.RegisterType[nodes.Struct[TransformAreaNode]](factory)
	refutil.RegisterType[nodes.Struct[OffsetNode]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode]](factory)
	generator.RegisterTypes(factory)
}
