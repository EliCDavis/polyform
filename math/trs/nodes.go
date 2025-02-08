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

	generator.RegisterTypes(factory)
}

type ArrayNode = nodes.Struct[[]TRS, ArrayNodeData]

type ArrayNodeData struct {
	Position nodes.NodeOutput[[]vector3.Float64]
	Scale    nodes.NodeOutput[[]vector3.Float64]
}

func (tnd ArrayNodeData) Process() ([]TRS, error) {
	positions := tnd.Position.Value()
	scales := tnd.Scale.Value()

	transforms := make([]TRS, len(positions))
	for i := 0; i < len(transforms); i++ {
		transforms[i] = New(
			positions[i],
			quaternion.Identity(),
			scales[i],
		)
	}

	return transforms, nil
}
