package meshops

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[TranslateAttribute3DNode]](factory)
	refutil.RegisterType[nodes.Struct[TranslateAttributeByPerlinNoise3DNode]](factory)
	refutil.RegisterType[nodes.Struct[RotateAttribute3DNode]](factory)
	refutil.RegisterType[nodes.Struct[CropAttribute3DNode]](factory)
	refutil.RegisterType[nodes.Struct[CenterAttribute3DNode]](factory)
	refutil.RegisterType[nodes.Struct[LaplacianSmoothNode]](factory)
	refutil.RegisterType[nodes.Struct[LaplacianSmoothImplicitWeldNode]](factory)
	refutil.RegisterType[nodes.Struct[SrgbToLinearNode]](factory)
	refutil.RegisterType[nodes.Struct[LinearToSRGBNode]](factory)

	refutil.RegisterType[nodes.Struct[CombineNode]](factory)

	refutil.RegisterType[nodes.Struct[SmoothNormalsNode]](factory)
	refutil.RegisterType[nodes.Struct[SmoothNormalsImplicitWeldNode]](factory)
	refutil.RegisterType[nodes.Struct[FlatNormalsNode]](factory)

	refutil.RegisterType[nodes.Struct[ScaleAttribute3DNode]](factory)
	refutil.RegisterType[nodes.Struct[ScaleAttributeAlongNormalNode]](factory)

	refutil.RegisterType[nodes.Struct[SliceAttributeByPlaneNode]](factory)
	refutil.RegisterType[nodes.Struct[FlipTriangleWindingNode]](factory)

	generator.RegisterTypes(factory)
}
