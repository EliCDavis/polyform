package repeat

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[MeshNode](factory)
	refutil.RegisterType[CircleNode](factory)
	refutil.RegisterType[SplineNode](factory)
	refutil.RegisterType[LineNode](factory)
	generator.RegisterTypes(factory)
}
