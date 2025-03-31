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

	minOpacity := -math.MaxFloat64
	maxOpacity := math.MaxFloat64
	minVolume := -math.MaxFloat64
	maxVolume := math.MaxFloat64

	if fnd.MinOpacity != nil {
		minOpacity = fnd.MinOpacity.Value()
	}

	if fnd.MaxOpacity != nil {
		maxOpacity = fnd.MaxOpacity.Value()
	}

	if fnd.MinVolume != nil {
		minVolume = fnd.MinVolume.Value()
	}

	if fnd.MaxVolume != nil {
		maxVolume = fnd.MaxVolume.Value()
	}

	m := fnd.Splat.Value()
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

	return nodes.NewStructOutput(meshops.RemovedUnreferencedVertices(m.SetIndices(indicesKept)))
}
