package parameter

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}

	refutil.RegisterTypeWithBuilder(factory, func() Value[float64] {
		return Value[float64]{
			Name: "Float",
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() Value[geometry.AABB] {
		return Value[geometry.AABB]{
			Name:         "Box",
			DefaultValue: geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]()),
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() Value[vector3.Float64] {
		return Value[vector3.Float64]{
			Name:         "Position",
			DefaultValue: vector3.Zero[float64](),
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() Value[coloring.WebColor] {
		return Value[coloring.WebColor]{
			Name:         "Color",
			DefaultValue: coloring.White(),
		}
	})

	return factory
}
