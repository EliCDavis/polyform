package gausops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
)

type FilterNode struct {
	Splat nodes.Output[modeling.Mesh]

	MinOpacity nodes.Output[float64] `description:"Splats less opaque than this are removed. With Max Opacity at or below it, opacity is not filtered."`
	MaxOpacity nodes.Output[float64] `description:"Splats more opaque than this are removed. At or below Min Opacity, opacity is not filtered."`
	MinVolume  nodes.Output[float64] `description:"Splats smaller than this are removed. With Max Volume at or below it, volume is not filtered."`
	MaxVolume  nodes.Output[float64] `description:"Splats larger than this are removed. At or below Min Volume, volume is not filtered."`
}

func (fnd FilterNode) Description() string {
	return "Removes gaussian splats whose opacity or volume falls outside the given range."
}

func (fnd FilterNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	if fnd.Splat == nil {
		out.Set(modeling.EmptyPointcloud())
		return
	}

	minOpacity := nodes.GetOutputValue(out, fnd.MinOpacity)
	maxOpacity := nodes.GetOutputValue(out, fnd.MaxOpacity)
	minVolume := nodes.GetOutputValue(out, fnd.MinVolume)
	maxVolume := nodes.GetOutputValue(out, fnd.MaxVolume)
	byOpacity := maxOpacity > minOpacity
	byVolume := maxVolume > minVolume

	m := nodes.GetOutputValue(out, fnd.Splat)
	opacity := m.Float1Attribute(modeling.OpacityAttribute)
	scale := m.Float3Attribute(modeling.ScaleAttribute)

	indicesKept := make([]int, 0)
	for i := 0; i < opacity.Len(); i++ {
		if byOpacity && (opacity.At(i) < minOpacity || opacity.At(i) > maxOpacity) {
			continue
		}

		exp := scale.At(i).Exp()
		length := exp.X() * exp.Y() * exp.Z()
		if byVolume && (length < minVolume || length > maxVolume) {
			continue
		}
		indicesKept = append(indicesKept, i)
	}

	out.Set(meshops.RemovedUnreferencedVertices(m.SetIndices(indicesKept)))
}
