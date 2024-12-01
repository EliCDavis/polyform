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

type LengthNode = nodes.StructNode[float64, LengthNodeData]

type LengthNodeData struct {
	Spline nodes.NodeOutput[Spline]
}

func (r LengthNodeData) Process() (float64, error) {

	if r.Spline == nil {
		return 0, nil
	}

	spline := r.Spline.Value()

	if spline == nil {
		return 0, nil
	}

	return spline.Length(), nil
}
