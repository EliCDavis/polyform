package vector2

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type Normalize struct {
	In nodes.LiftedPort[vector2.Float64]
}

func (cn Normalize) Description() string {
	return "Scales a vector to unit length."
}

func (cn Normalize) Normalized(out *nodes.Lifted[vector2.Float64]) {
	nodes.Zip1(out, cn.In, func(v vector2.Float64) vector2.Float64 {
		if v.Length() == 0 {
			return vector2.Zero[float64]()
		}
		return v.Normalized()
	})
}

func (cn Normalize) NormalizeDescription() string {
	return "Returns the input vector scaled to have a length of 1. (0,0) is returned if no vector is provided"
}

// ============================================================================

type NormalizeArray struct {
	In nodes.Output[[]vector2.Float64]
}

func (cn NormalizeArray) Description() string {
	return "Scales a whole array of vectors by the length of the longest one, so their relative lengths survive."
}

func (cn NormalizeArray) Global(out *nodes.StructOutput[[]vector2.Float64]) {
	if cn.In == nil {
		return
	}

	in := nodes.GetOutputValue(out, cn.In)
	if len(in) == 0 {
		return
	}

	maxMagnitude := 0.
	for _, v := range in {
		maxMagnitude = max(maxMagnitude, v.Length())
	}

	if maxMagnitude == 0 {
		out.CaptureError(fmt.Errorf("all vector data has a magnitude of 0"))
		out.Set(in)
		return
	}

	arr := make([]vector2.Float64, len(in))
	for i, v := range in {
		arr[i] = v.DivByConstant(maxMagnitude)
	}
	out.Set(arr)
}

func (cn NormalizeArray) GlobalDescription() string {
	return "Scales each vector by the inverse of the magnitude of the longest vector"
}
