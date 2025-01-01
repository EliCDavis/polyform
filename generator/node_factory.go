package generator

import (
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/meshops/gausops"
	"github.com/EliCDavis/polyform/refutil"
)

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}

	// refutil.RegisterTypeWithBuilder(factory, func() ParameterNode[float64] {
	// 	return ParameterNode[float64]{
	// 		Name: "Float",
	// 	}
	// })

	// refutil.RegisterTypeWithBuilder(factory, func() parameter.Value[geometry.AABB] {
	// 	return parameter.Value[geometry.AABB]{
	// 		Name:         "Box",
	// 		DefaultValue: geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]()),
	// 	}
	// })

	// refutil.RegisterTypeWithBuilder(factory, func() ParameterNode[vector3.Float64] {
	// 	return ParameterNode[vector3.Float64]{
	// 		Name:         "Position",
	// 		DefaultValue: vector3.Zero[float64](),
	// 	}
	// })

	// refutil.RegisterTypeWithBuilder(factory, func() ParameterNode[coloring.WebColor] {
	// 	return ParameterNode[coloring.WebColor]{
	// 		Name:         "Color",
	// 		DefaultValue: coloring.White(),
	// 	}
	// })

	return factory.Combine(
		meshops.Nodes(),
		gausops.Nodes(),
		parameter.Nodes(),
	)
}
