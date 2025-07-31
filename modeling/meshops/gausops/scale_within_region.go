package gausops

import (
	"errors"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type ScaleWithinRegionNode = nodes.Struct[ScaleWithinRegionNodeData]

type ScaleWithinRegionNodeData struct {
	Mesh     nodes.Output[modeling.Mesh]
	Scale    nodes.Output[float64]
	Radius   nodes.Output[float64]
	Position nodes.Output[vector3.Float64]
}

func (swrnd ScaleWithinRegionNodeData) Out() nodes.StructOutput[modeling.Mesh] {
	if swrnd.Mesh == nil {
		return nodes.NewStructOutput(modeling.EmptyPointcloud())
	}

	out := nodes.StructOutput[modeling.Mesh]{}
	m := nodes.GetOutputValue(&out, swrnd.Mesh)

	if !m.HasFloat3Attribute(modeling.PositionAttribute) || !m.HasFloat3Attribute(modeling.ScaleAttribute) {
		out.CaptureError(errors.New("requires mesh with position and scaling data"))
		return out
	}

	posData := m.Float3Attribute(modeling.PositionAttribute)
	scaleData := m.Float3Attribute(modeling.ScaleAttribute)
	count := posData.Len()

	newPos := make([]vector3.Float64, count)
	newScale := make([]vector3.Float64, count)

	baloonPos := nodes.TryGetOutputValue(&out, swrnd.Position, vector3.Zero[float64]())
	baloonRadius := nodes.TryGetOutputValue(&out, swrnd.Radius, 1)
	baloonStrength := nodes.TryGetOutputValue(&out, swrnd.Scale, 1)

	for i := range count {
		curPos := posData.At(i)
		curScale := scaleData.At(i)
		dir := curPos.Sub(baloonPos)
		len := dir.Length()

		if len <= baloonRadius {
			newPos[i] = baloonPos.Add(dir.Scale(baloonStrength))
			newScale[i] = curScale.Exp().Scale(baloonStrength).Log()
		} else {
			newPos[i] = curPos
			newScale[i] = curScale
		}
	}
	out.Set(m.
		SetFloat3Attribute(modeling.PositionAttribute, newPos).
		SetFloat3Attribute(modeling.ScaleAttribute, newScale))

	return out
}
