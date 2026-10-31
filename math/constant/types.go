package constant

import (
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[Vector3[float64]](factory)
	refutil.RegisterType[Vector3[int]](factory)
	refutil.RegisterType[Pi](factory)
	refutil.RegisterType[Quaternion](factory)

	registry.Register(factory)
}
