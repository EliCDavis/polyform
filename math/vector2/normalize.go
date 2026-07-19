package vector2

import (
	"fmt"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type NormalizeArray struct {
	In nodes.Output[[]vector2.Float64]
}

func (cn NormalizeArray) Local(out *nodes.StructOutput[[]vector2.Float64]) {
	if cn.In == nil {
		return
	}

	in := nodes.GetOutputValue(out, cn.In)
	arr := make([]vector2.Float64, len(in))
	for i, v := range in {
		arr[i] = v.Normalized()
	}
	out.Set(arr)
}

func (cn NormalizeArray) LocalDescription() string {
	return "Normalizes each component of the array"
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
