package gausops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
)

type FilterNode = nodes.Struct[FilterNodeData]

type FilterNodeData struct {
	Splat nodes.Output[modeling.Mesh]

	MinOpacity nodes.Output[float64]
	MaxOpacity nodes.Output[float64]
	MinVolume  nodes.Output[float64]
	MaxVolume  nodes.Output[float64]
}

func (fnd FilterNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if fnd.Splat == nil {
		return nodes.NewStructOutput(modeling.EmptyPointcloud())
	}

	out := nodes.StructOutput[modeling.Mesh]{}
	minOpacity := nodes.TryGetOutputValue(&out, fnd.MinOpacity, -math.MaxFloat64)
	maxOpacity := nodes.TryGetOutputValue(&out, fnd.MaxOpacity, math.MaxFloat64)
	minVolume := nodes.TryGetOutputValue(&out, fnd.MinVolume, -math.MaxFloat64)
	maxVolume := nodes.TryGetOutputValue(&out, fnd.MaxVolume, math.MaxFloat64)

	m := nodes.GetOutputValue(out, fnd.Splat)
	opacity := m.Float1Attribute(modeling.OpacityAttribute)
	scale := m.Float3Attribute(modeling.ScaleAttribute)

	indicesKept := make([]int, 0)
	for i := 0; i < opacity.Len(); i++ {
		if opacity.At(i) < minOpacity || opacity.At(i) > maxOpacity {
			continue
		}

		exp := scale.At(i).Exp()
		length := exp.X() * exp.Y() * exp.Z()
		if length < minVolume || length > maxVolume {
			continue
		}
		indicesKept = append(indicesKept, i)
	}

	out.Set(meshops.RemovedUnreferencedVertices(m.SetIndices(indicesKept)))
	return out
}
