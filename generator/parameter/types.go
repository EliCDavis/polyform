package parameter

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterTypeWithBuilder(factory, func() Int { return Int{} })
	refutil.RegisterTypeWithBuilder(factory, func() Float64 { return Float64{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector3 { return Vector3{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector2 { return Vector2{} })
	refutil.RegisterTypeWithBuilder(factory, func() Bool { return Bool{} })
	refutil.RegisterTypeWithBuilder(factory, func() String { return String{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector3Array { return Vector3Array{} })

	refutil.RegisterTypeWithBuilder(factory, func() AABB {
		return AABB{
			DefaultValue: geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]()),
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() Color {
		return Color{
			DefaultValue: coloring.White(),
		}
	})

	generator.RegisterTypes(factory)
}
