package meshops

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[TranslateAttribute3DNode](factory)
	refutil.RegisterType[nodes.Struct[TranslateAttributeByPerlinNoise3DNode]](factory)
	refutil.RegisterType[CropAttribute3DNode](factory)
	refutil.RegisterType[CenterAttribute3DNode](factory)
	refutil.RegisterType[LaplacianSmoothNode](factory)

	refutil.RegisterType[CombineNode](factory)

	refutil.RegisterType[SmoothNormalsNode](factory)
	refutil.RegisterType[SmoothNormalsImplicitWeldNode](factory)
	refutil.RegisterType[FlatNormalsNode](factory)

	refutil.RegisterType[ScaleAttribute3DNode](factory)
	refutil.RegisterType[ScaleAttributeAlongNormalNode](factory)

	generator.RegisterTypes(factory)
}
