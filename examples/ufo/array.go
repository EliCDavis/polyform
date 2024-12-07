package main

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type TRSNodeData struct {
	Position nodes.NodeOutput[[]vector3.Float64]
	Scale    nodes.NodeOutput[[]vector3.Float64]
}

type TRSNode = nodes.StructNode[[]trs.TRS, TRSNodeData]

func (tnd TRSNodeData) Process() ([]trs.TRS, error) {
	positions := tnd.Position.Value()
	scales := tnd.Scale.Value()

	transforms := make([]trs.TRS, len(positions))
	for i := 0; i < len(transforms); i++ {
		transforms[i] = trs.New(
			positions[i],
			quaternion.Identity(),
			scales[i],
		)
	}

	return transforms, nil
}
