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
	refutil.RegisterType[nodes.Struct[UniformNode[coloring.WebColor]]](factory)

	refutil.RegisterType[nodes.Struct[FromArrayNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[coloring.WebColor]]](factory)

	refutil.RegisterType[nodes.Struct[SelectNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[coloring.WebColor]]](factory)

	refutil.RegisterType[nodes.Struct[CompareValueTextureNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[LinearGradientNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[coloring.WebColor]]](factory)

	refutil.RegisterType[nodes.Struct[ApplyMaskNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[ApplyMaskNode[coloring.WebColor]]](factory) // <-- Not sure how people are to connect with this one

	refutil.RegisterType[nodes.Struct[ColorToImageNode]](factory)
	refutil.RegisterType[nodes.Struct[FloatToImageNode]](factory)
	refutil.RegisterType[nodes.Struct[SeamlessPerlinNode]](factory)
	refutil.RegisterType[nodes.Struct[PerlinNode]](factory)

	refutil.RegisterType[nodes.Struct[AddFloat1Node]](factory)
	refutil.RegisterType[nodes.Struct[AddFloat2Node]](factory)
	refutil.RegisterType[nodes.Struct[AddFloat3Node]](factory)
	refutil.RegisterType[nodes.Struct[AddFloat4Node]](factory)

	generator.RegisterTypes(factory)
}
