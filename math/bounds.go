package math

import (
	"errors"
	"slices"

	"github.com/EliCDavis/polyform/nodes"
)

type BoundsNode struct {
	Values  nodes.Output[[]float64] `description:"Values the range must contain."`
	Padding nodes.Output[float64]   `description:"Distance added past the smallest and largest value."`
}

func (n BoundsNode) Description() string {
	return "Smallest range containing every value, grown by a padding on both ends."
}

func (n BoundsNode) bounds(recorder nodes.ExecutionRecorder) (float64, float64) {
	values := nodes.TryGetOutputValue(recorder, n.Values, nil)
	if len(values) == 0 {
		recorder.CaptureError(errors.New("no values to bound"))
		return 0, 0
	}
	padding := nodes.TryGetOutputValue(recorder, n.Padding, 0)
	return slices.Min(values) - padding, slices.Max(values) + padding
}

func (n BoundsNode) Min(out *nodes.StructOutput[float64]) {
	min, _ := n.bounds(out)
	out.Set(min)
}

func (n BoundsNode) Max(out *nodes.StructOutput[float64]) {
	_, max := n.bounds(out)
	out.Set(max)
}

func (n BoundsNode) Center(out *nodes.StructOutput[float64]) {
	min, max := n.bounds(out)
	out.Set((min + max) / 2)
}

func (n BoundsNode) Size(out *nodes.StructOutput[float64]) {
	min, max := n.bounds(out)
	out.Set(max - min)
}
