package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector3"
)

func Box(pos vector3.Float64, bounds vector3.Float64) sample.Vec3ToFloat {
	halfBounds := bounds.Scale(0.5)
	// It's best to watch the video to understand
	// https://www.youtube.com/watch?v=62-pRVZuS5c
	return func(v vector3.Float64) float64 {
		reorient := v.Sub(pos)
		q := reorient.Abs().Sub(halfBounds)

		inside := math.Min(math.Max(q.X(), math.Max(q.Y(), q.Z())), 0)
		xLength := math.Pow(math.Max(0, q.X()), 2.)
		yLength := math.Pow(math.Max(0, q.Y()), 2.)
		zLength := math.Pow(math.Max(0, q.Z()), 2.)
		return math.Sqrt(xLength+yLength+zLength) + inside
	}
}
