package meshops

import (
	"github.com/EliCDavis/polyform/refutil"
)

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[TranslateAttribute3DNode](factory)
	refutil.RegisterType[CropAttribute3DNode](factory)
	refutil.RegisterType[LaplacianSmoothNode](factory)
	refutil.RegisterType[SmoothNormalsNode](factory)
	refutil.RegisterType[ScaleAttribute3DNode](factory)
	refutil.RegisterType[CombineNode](factory)
	refutil.RegisterType[SmoothNormalsImplicitWeldNode](factory)

	return factory
}
