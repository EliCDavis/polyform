package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CirclePointsNode = nodes.Struct[[]vector3.Float64, CirclePointsNodeData]

type CirclePointsNodeData struct {
	Count  nodes.NodeOutput[int]
	Radius nodes.NodeOutput[float64]
}

func (cpnd CirclePointsNodeData) Process() ([]vector3.Float64, error) {
	count := 0
	radius := 1.

	if cpnd.Count != nil {
		count = cpnd.Count.Value()
	}

	if cpnd.Radius != nil {
		radius = cpnd.Radius.Value()
	}

	return CirclePoints(count, radius), nil
}

func CirclePoints(count int, radius float64) []vector3.Float64 {
	angleIncrement := (1.0 / float64(count)) * 2.0 * math.Pi
	final := make([]vector3.Float64, count)

	for i := 0; i < count; i++ {
		angle := angleIncrement * float64(i)
		final[i] = vector3.New(math.Cos(angle)*radius, 0, math.Sin(angle)*radius)
	}

	return final
}

func Circle(times int, radius float64) []trs.TRS {
	angleIncrement := (1.0 / float64(times)) * 2.0 * math.Pi

	transforms := make([]trs.TRS, times)

	for i := 0; i < times; i++ {
		angle := angleIncrement * float64(i)

		pos := vector3.New(math.Cos(angle), 0, math.Sin(angle)).Scale(radius)
		rot := quaternion.FromTheta(angle-(math.Pi/2), vector3.Down[float64]())

		transforms[i] = trs.New(pos, rot, vector3.One[float64]())
	}

	return transforms
}

type CircleNode = nodes.Struct[[]trs.TRS, CircleNodeData]

type CircleNodeData struct {
	Radius nodes.NodeOutput[float64]
	Times  nodes.NodeOutput[int]
}

func (r CircleNodeData) Process() ([]trs.TRS, error) {
	times := 0
	radius := 0.

	if r.Times != nil {
		times = r.Times.Value()
	}

	if r.Radius != nil {
		radius = r.Radius.Value()
	}

	return Circle(times, radius), nil
}
