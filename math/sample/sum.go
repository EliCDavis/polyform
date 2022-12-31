package sample

import "github.com/EliCDavis/vector"

func SumVec3ToFloat(v32fs ...Vec3ToFloat) Vec3ToFloat {
	if len(v32fs) == 0 {
		panic("can not create a sum without any functions")
	}

	if len(v32fs) == 1 {
		return v32fs[0]
	}

	return func(f vector.Vector3) float64 {
		result := v32fs[0](f)
		for i := 1; i < len(v32fs); i++ {
			result += v32fs[i](f)
		}
		return result
	}
}
