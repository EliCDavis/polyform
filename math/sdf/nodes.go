package sdf

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[TranslateNode]](factory)
	refutil.RegisterType[nodes.Struct[TransformNode]](factory)

	refutil.RegisterType[nodes.Struct[UnionNode]](factory)
	refutil.RegisterType[nodes.Struct[IntersectionNode]](factory)
	refutil.RegisterType[nodes.Struct[SubtractionNode]](factory)
	refutil.RegisterType[nodes.Struct[MirrorNode]](factory)

	refutil.RegisterType[nodes.Struct[CubeNode]](factory)
	refutil.RegisterType[nodes.Struct[RoundCubeNode]](factory)
	refutil.RegisterType[nodes.Struct[LineNode]](factory)
	refutil.RegisterType[nodes.Struct[PlaneNode]](factory)
	refutil.RegisterType[nodes.Struct[RoundedConeNode]](factory)
	refutil.RegisterType[nodes.Struct[RoundedCylinderNode]](factory)
	refutil.RegisterType[nodes.Struct[SphereNode]](factory)

	generator.RegisterTypes(factory)
}
