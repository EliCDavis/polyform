package curves

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type Curve interface {
	At(t float64) vector3.Float64
}

type Spline interface {
	Length() float64
	At(distance float64) vector3.Float64
	Dir(distance float64) vector3.Float64
}

type LengthNode = nodes.Struct[LengthNodeData]

type LengthNodeData struct {
	Spline nodes.Output[Spline]
}

func (r LengthNodeData) Out() nodes.StructOutput[float64] {
	if r.Spline == nil {
		return nodes.NewStructOutput(0.)
	}

	spline := r.Spline.Value()

	if spline == nil {
		return nodes.NewStructOutput(0.)
	}

	return nodes.NewStructOutput(spline.Length())
}
