package texturing

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[UniformNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[UniformNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[UniformNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[UniformNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[UniformNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[FromArrayNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[SelectNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[CompareValueTextureNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[LinearGradientNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[ApplyMaskNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[coloring.Color]]](factory)

	refutil.RegisterType[nodes.Struct[ColorToImageNode]](factory)
	refutil.RegisterType[nodes.Struct[FloatToImageNode]](factory)
	refutil.RegisterType[nodes.Struct[SeamlessPerlinNode]](factory)
	refutil.RegisterType[nodes.Struct[PerlinNode]](factory)
	refutil.RegisterType[nodes.Struct[DebugUVNode]](factory)
	refutil.RegisterType[nodes.Struct[NoiseNode]](factory)

	refutil.RegisterType[nodes.Struct[AddFloat1Node]](factory)
	refutil.RegisterType[nodes.Struct[AddFloat2Node]](factory)
	refutil.RegisterType[nodes.Struct[AddFloat3Node]](factory)
	refutil.RegisterType[nodes.Struct[AddFloat4Node]](factory)

	refutil.RegisterType[nodes.Struct[ScaleFloat1Node]](factory)
	refutil.RegisterType[nodes.Struct[ScaleFloat2Node]](factory)
	refutil.RegisterType[nodes.Struct[ScaleFloat3Node]](factory)
	refutil.RegisterType[nodes.Struct[ScaleFloat4Node]](factory)

	refutil.RegisterType[nodes.Struct[MultiplyFloat1Node]](factory)

	generator.RegisterTypes(factory)
}
