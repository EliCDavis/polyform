package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

func Union(fields ...sample.Vec3ToFloat) sample.Vec3ToFloat {
	if len(fields) == 0 {
		panic("no fields to union")
	}

	if len(fields) == 1 {
		return fields[0]
	}

	return func(v vector3.Float64) float64 {
		min := fields[0](v)
		for i := 1; i < len(fields); i++ {
			min = math.Min(min, fields[i](v))
		}
		return min
	}
}

func Intersect(fields ...sample.Vec3ToFloat) sample.Vec3ToFloat {
	if len(fields) == 0 {
		panic("no fields to intersect")
	}

	if len(fields) == 1 {
		return fields[0]
	}

	return func(v vector3.Float64) float64 {
		max := fields[0](v)
		for i := 1; i < len(fields); i++ {
			max = math.Max(max, fields[i](v))
		}
		return max
	}
}
