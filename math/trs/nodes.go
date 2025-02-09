package trs

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector/vector3"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ArrayNode](factory)
	refutil.RegisterType[NewNode](factory)

	generator.RegisterTypes(factory)
}

type NewNode = nodes.Struct[TRS, NewNodeData]

type NewNodeData struct {
	Position nodes.NodeOutput[vector3.Float64]
	Scale    nodes.NodeOutput[vector3.Float64]
	Rotation nodes.NodeOutput[quaternion.Quaternion]
}

func (tnd NewNodeData) Process() (TRS, error) {
	return New(
		nodes.TryGetOutputValue(tnd.Position, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(tnd.Rotation, quaternion.Identity()),
		nodes.TryGetOutputValue(tnd.Scale, vector3.One[float64]()),
	), nil
}

type ArrayNode = nodes.Struct[[]TRS, ArrayNodeData]

type ArrayNodeData struct {
	Position nodes.NodeOutput[[]vector3.Float64]
	Scale    nodes.NodeOutput[[]vector3.Float64]
	Rotation nodes.NodeOutput[[]quaternion.Quaternion]
}

func (tnd ArrayNodeData) Process() ([]TRS, error) {
	positions := nodes.TryGetOutputValue(tnd.Position, nil)
	rotations := nodes.TryGetOutputValue(tnd.Rotation, nil)
	scales := nodes.TryGetOutputValue(tnd.Scale, nil)

	transforms := make([]TRS, max(len(positions), len(rotations), len(scales)))
	for i := 0; i < len(transforms); i++ {
		p := vector3.Zero[float64]()
		r := quaternion.Identity()
		s := vector3.One[float64]()

		if i < len(positions) {
			p = positions[i]
		}

		if i < len(rotations) {
			r = rotations[i]
		}

		if i < len(scales) {
			s = scales[i]
		}

		transforms[i] = New(p, r, s)
	}

	return transforms, nil
}
