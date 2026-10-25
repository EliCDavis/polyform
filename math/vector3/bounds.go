package vector3

import (
	"errors"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type BoundsNode struct {
	Points  nodes.Output[[]vector3.Float64] `description:"Points the box must contain."`
	Padding nodes.Output[float64]           `description:"Distance added to every side of the box."`
}

func (n BoundsNode) Description() string {
	return "Smallest axis-aligned box containing every point, grown by a padding on every side."
}

func (n BoundsNode) bounds(recorder nodes.ExecutionRecorder) (vector3.Float64, vector3.Float64) {
	points := nodes.TryGetOutputValue(recorder, n.Points, nil)
	if len(points) == 0 {
		recorder.CaptureError(errors.New("no points to bound"))
		return vector3.Zero[float64](), vector3.Zero[float64]()
	}

	min, max := points[0], points[0]
	for _, p := range points[1:] {
		min = vector3.Min(min, p)
		max = vector3.Max(max, p)
	}

	padding := nodes.TryGetOutputValue(recorder, n.Padding, 0)
	pad := vector3.New(padding, padding, padding)
	return min.Sub(pad), max.Add(pad)
}

func (n BoundsNode) Min(out *nodes.StructOutput[vector3.Float64]) {
	min, _ := n.bounds(out)
	out.Set(min)
}

func (n BoundsNode) Max(out *nodes.StructOutput[vector3.Float64]) {
	_, max := n.bounds(out)
	out.Set(max)
}

func (n BoundsNode) Center(out *nodes.StructOutput[vector3.Float64]) {
	min, max := n.bounds(out)
	out.Set(min.Midpoint(max))
}

func (n BoundsNode) Size(out *nodes.StructOutput[vector3.Float64]) {
	min, max := n.bounds(out)
	out.Set(max.Sub(min))
}

func (n BoundsNode) AABB(out *nodes.StructOutput[geometry.AABB]) {
	min, max := n.bounds(out)
	box := geometry.NewEmptyAABB()
	box.SetMinMax(min, max)
	out.Set(box)
}
