package parameter

import (
	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[File](factory)
	refutil.RegisterType[Image](factory)

	refutil.RegisterTypeWithBuilder(factory, func() Int { return Int{} })
	refutil.RegisterTypeWithBuilder(factory, func() Float64 { return Float64{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector3 { return Vector3{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector2 { return Vector2{} })
	refutil.RegisterTypeWithBuilder(factory, func() Bool { return Bool{} })
	refutil.RegisterTypeWithBuilder(factory, func() String { return String{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector4 { return Vector4{} })
	refutil.RegisterTypeWithBuilder(factory, func() Float64Array { return Float64Array{} })
	refutil.RegisterTypeWithBuilder(factory, func() IntArray { return IntArray{} })
	refutil.RegisterTypeWithBuilder(factory, func() StringArray { return StringArray{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector2Array { return Vector2Array{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector2IntArray { return Vector2IntArray{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector3Array { return Vector3Array{} })
	refutil.RegisterTypeWithBuilder(factory, func() Vector3IntArray { return Vector3IntArray{} })

	refutil.RegisterTypeWithBuilder(factory, func() AABB {
		return AABB{
			CurrentValue: geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]()),
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() Color {
		return Color{
			CurrentValue: coloring.White(),
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() ColorArray { return ColorArray{} })
	refutil.RegisterTypeWithBuilder(factory, func() ColorGradient {
		return ColorGradient{
			CurrentValue: coloring.NewGradientColor(
				coloring.GradientKey[coloring.Color]{Time: 0, Value: coloring.Black()},
				coloring.GradientKey[coloring.Color]{Time: 1, Value: coloring.White()},
			),
		}
	})

	refutil.RegisterTypeWithBuilder(factory, func() Quaternion {
		return Quaternion{CurrentValue: quaternion.Identity()}
	})
	refutil.RegisterTypeWithBuilder(factory, func() TRS {
		return TRS{CurrentValue: trs.Identity()}
	})
	refutil.RegisterTypeWithBuilder(factory, func() TRSArray { return TRSArray{} })

	generator.RegisterTypes(factory)
}
