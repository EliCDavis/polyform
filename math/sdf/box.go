package sdf

import (
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Box(position, bounds vector3.Float64) sample.Vec3ToFloat {
	halfBounds := bounds.Scale(0.5)
	// It's best to watch the video to understand
	// https://www.youtube.com/watch?v=62-pRVZuS5c
	return func(v vector3.Float64) float64 {
		reorient := v.Sub(position)
		q := reorient.Abs().Sub(halfBounds)

		inside := math.Min(q.MaxComponent(), 0)
		return vector3.Max(q, vector3.Zero[float64]()).Length() + inside
	}
}

type CubeNode struct {
	Position nodes.Output[vector3.Float64] `description:"Center of the box. Defaults to the origin."`
	Size     nodes.Output[vector3.Float64] `description:"Full width/height/depth of the box. Defaults to (1, 1, 1)."`
}

func (cn CubeNode) Description() string {
	return "An axis-aligned box."
}

func (cn CubeNode) Field(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	out.Set(Box(
		nodes.TryGetOutputValue(out, cn.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, cn.Size, vector3.One[float64]()),
	))
}
