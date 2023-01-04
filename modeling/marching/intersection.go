package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector"
)

func FieldIntersection(fields ...sample.Vec3ToFloat) sample.Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		value := 0.

		for _, field := range fields {
			fieldValue := field(v)
			if fieldValue <= 0 {
				return 0
			}
			value += fieldValue
		}

		return value
	}
}
