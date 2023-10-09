package marching

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

func Subtract(base, subtraction Field) Field {
	float1Functions := make(map[string]sample.Vec3ToFloat)
	float2Functions := make(map[string]sample.Vec3ToVec2)
	float3Functions := make(map[string]sample.Vec3ToVec3)

	newDomain := geometry.NewEmptyAABB()
	newDomain.EncapsulateBounds(base.Domain)
	newDomain.EncapsulateBounds(subtraction.Domain)

	for atr, f := range subtraction.Float1Functions {
		baseFun := base.Float1Functions[atr]
		float1Functions[atr] = func(v vector3.Float64) float64 {
			return math.Max(baseFun(v), -f(v))

			// inBase := base.Domain.Contains(v)
			// inSub := subtraction.Domain.Contains(v)

			// if inBase && inSub {
			// return math.Max(baseFun(v), -f(v))
			// }

			// if inSub {
			// return -f(v)
			// }

			// return baseFun(v)
		}
	}

	// for atr, f := range field.Float2Functions {
	// 	float2Functions[atr] = func(v vector3.Float64) vector2.Float64 {
	// 		return f(newV(v))
	// 	}
	// }

	// for atr, f := range field.Float3Functions {
	// 	float3Functions[atr] = func(v vector3.Float64) vector3.Float64 {
	// 		return f(newV(v))
	// 	}
	// }

	return Field{
		Domain:          newDomain,
		Float1Functions: float1Functions,
		Float2Functions: float2Functions,
		Float3Functions: float3Functions,
	}
}
