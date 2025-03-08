package gausops

import (
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

	m := swrnd.Mesh.Value()

	posData := m.Float3Attribute(modeling.PositionAttribute)
	scaleData := m.Float3Attribute(modeling.ScaleAttribute)
	count := posData.Len()

	newPos := make([]vector3.Float64, count)
	newScale := make([]vector3.Float64, count)

	baloonPos := swrnd.Position.Value()
	baloonRadius := swrnd.Radius.Value()
	baloonStrength := swrnd.Scale.Value()

	for i := 0; i < count; i++ {
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

	return nodes.NewStructOutput(m.
		SetFloat3Attribute(modeling.PositionAttribute, newPos).
		SetFloat3Attribute(modeling.ScaleAttribute, newScale))
}
