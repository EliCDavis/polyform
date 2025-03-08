package repeat

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CirclePointsNode = nodes.Struct[CirclePointsNodeData]

type CirclePointsNodeData struct {
	Count  nodes.Output[int]
	Radius nodes.Output[float64]
}

func (cpnd CirclePointsNodeData) Out() nodes.StructOutput[[]vector3.Float64] {
	count := 0
	radius := 1.

	if cpnd.Count != nil {
		count = cpnd.Count.Value()
	}

	if cpnd.Radius != nil {
		radius = cpnd.Radius.Value()
	}

	return nodes.NewStructOutput(CirclePoints(count, radius))
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

type CircleNode = nodes.Struct[CircleNodeData]

type CircleNodeData struct {
	Radius nodes.Output[float64]
	Times  nodes.Output[int]
}

func (r CircleNodeData) Out() nodes.StructOutput[[]trs.TRS] {
	times := 0
	radius := 0.

	if r.Times != nil {
		times = r.Times.Value()
	}

	if r.Radius != nil {
		radius = r.Radius.Value()
	}

	return nodes.NewStructOutput(Circle(times, radius))
}
