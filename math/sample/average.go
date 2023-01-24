package sample

import (
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func AverageVec3ToFloat(funcs ...Vec3ToFloat) Vec3ToFloat {
	if len(funcs) == 0 {
		panic("can not average the result of 0 functions")
	}

	if len(funcs) == 1 {
		return funcs[0]
	}

	return func(v vector3.Float64) float64 {
		val := 0.

		for _, f := range funcs {
			val += f(v)
		}

		return val / float64(len(funcs))
	}
}

func AverageVec3ToVec2(funcs ...Vec3ToVec2) Vec3ToVec2 {
	if len(funcs) == 0 {
		panic("can not average the result of 0 functions")
	}

	if len(funcs) == 1 {
		return funcs[0]
	}

	return func(v vector3.Float64) vector2.Float64 {
		val := vector2.Zero[float64]()

		for _, f := range funcs {
			val = val.Add(f(v))
		}

		return val.DivByConstant(float64(len(funcs)))
	}
}

func AverageVec3ToVec3(funcs ...Vec3ToVec3) Vec3ToVec3 {
	if len(funcs) == 0 {
		panic("can not average the result of 0 functions")
	}

	if len(funcs) == 1 {
		return funcs[0]
	}

	return func(v vector3.Float64) vector3.Float64 {
		val := vector3.Zero[float64]()

		for _, f := range funcs {
			val = val.Add(f(v))
		}

		return val.DivByConstant(float64(len(funcs)))
	}
}
