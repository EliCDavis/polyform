package texturing

import (
	"image/color"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[TextureNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[TextureNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[TextureNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[TextureNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[TextureNode[color.Color]]](factory)

	refutil.RegisterType[nodes.Struct[FromArrayNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[FromArrayNode[color.Color]]](factory)

	refutil.RegisterType[nodes.Struct[SelectNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[bool]]](factory)
	refutil.RegisterType[nodes.Struct[SelectNode[color.Color]]](factory)

	refutil.RegisterType[nodes.Struct[CompareValueTextureNode[float64]]](factory)

	refutil.RegisterType[nodes.Struct[LinearGradientNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[vector2.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[vector3.Float64]]](factory)
	refutil.RegisterType[nodes.Struct[LinearGradientNode[color.Color]]](factory)

	refutil.RegisterType[nodes.Struct[ColorToImageNode]](factory)

	generator.RegisterTypes(factory)
}
