package gausops

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
)

func Filter(m modeling.Mesh, minOpacity, maxOpacity, minVolume, maxVolume float64) modeling.Mesh {
	byOpacity := maxOpacity > minOpacity
	byVolume := maxVolume > minVolume

	opacity := m.Float1Attribute(modeling.OpacityAttribute)
	scale := m.Float3Attribute(modeling.ScaleAttribute)

	kept := meshops.SelectIndices(opacity.Len(), func(i int) bool {
		if byOpacity && (opacity.At(i) < minOpacity || opacity.At(i) > maxOpacity) {
			return false
		}

		exp := scale.At(i).Exp()
		volume := exp.X() * exp.Y() * exp.Z()
		return !(byVolume && (volume < minVolume || volume > maxVolume))
	})

	return meshops.GatherPoints(m, kept)
}

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

	out.Set(Filter(
		nodes.GetOutputValue(out, fnd.Splat),
		nodes.GetOutputValue(out, fnd.MinOpacity),
		nodes.GetOutputValue(out, fnd.MaxOpacity),
		nodes.GetOutputValue(out, fnd.MinVolume),
		nodes.GetOutputValue(out, fnd.MaxVolume),
	))
}
